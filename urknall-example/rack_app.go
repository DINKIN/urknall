package main

import (
	"bytes"
	"github.com/dynport/dgtk/pubsub"
	"github.com/dynport/urknall"
	"github.com/dynport/urknall/cmd"
	"github.com/dynport/urknall/fw"
	"github.com/dynport/urknall/pkg/nginx"
	"github.com/dynport/urknall/pkg/ruby"
	"log"
	"os"
	"text/template"
)

var ps = pubsub.New()

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("USAGE: %s <IP>", os.Args[0])
	}

	// urknall does not do any logging publishes notifications for specific events to which
	// one can subscribe by registering a pubsub from github.com/dynport/dgtk/pubsub exchange with urknall.RegisterPubSub()
	// but there is also a default StdoutLogger() which can be activated like this:
	l, e := urknall.OpenStdoutLogger()
	if e != nil {
		log.Fatal(e.Error())
	}
	defer l.Close()

	// condfigure the host to provision
	host := &urknall.Host{
		IP: os.Args[1],
	}

	// upgrade all system packages
	host.AddCommands("package_upgrade", cmd.UpdatePackages())

	// use some standard packages
	nx := nginx.New("1.4.4")
	host.AddPackage("ruby", nx)

	rb := ruby.New("2.0.0-p353")
	host.AddPackage("nginx", rb)

	// install some custom system commands
	host.AddCommands("packages", cmd.InstallPackages("ngrep", "dnsutils", "whois"))

	// firewall setup: allow inbound http and https (currently ssh (22) is also allowed by default)
	host.Firewall = fw.Firewall{
		{
			Description: "Allow incoming http",
			Chain:       fw.ChainInput,
			Protocol:    fw.ProtocolTcp,
			Destination: &fw.Target{Port: fw.PortHttp},
		},
		{
			Description: "Allow incoming https",
			Chain:       fw.ChainInput,
			Protocol:    fw.ProtocolTcp,
			Destination: &fw.Target{Port: fw.PortHttps},
		},
	}

	app := &App{
		RubyInstallPath:  rb.InstallPath(),
		NginxInstallPath: nx.InstallPath(),
		SocketPath:       "/tmp/rack.socket",
		CurrentPath:      "/app/current",
	}
	host.AddPackage("app", app)

	// provision the host
	e = host.Provision(nil)
	if e != nil {
		log.Fatal(e.Error())
	}
}

type App struct {
	RubyInstallPath  string `urknall:"required=true"`
	NginxInstallPath string `urknall:"required=true"`
	SocketPath       string `urknall:"required=true"`
	CurrentPath      string `urknall:"required=true"`
}

func (app *App) Package(r *urknall.Runlist) {
	r.Add(
		"{{ .RubyInstallPath }}/bin/gem install puma --no-ri --no-rdoc", // is executed as plain bash command
		cmd.Mkdir("/app", "root", 0755),
		cmd.Mkdir("/app/public", "root", 0755),
		cmd.WriteFile(app.NginxInstallPath+"/conf/nginx.conf", nginxConfig, "root", 0644),
		"{ service nginx status && service nginx restart; } || service nginx start",
		cmd.WriteFile("/etc/init/app.conf", upstart, "root", 0755),
		cmd.WriteFile("/app/config.ru", configRu, "root", 0644),
		"{ service app status && service app restart; } || service app start",
	)
}

func mustRenderTemplate(src string, i interface{}) string {
	buf := &bytes.Buffer{}
	e := template.Must(template.New("nginx_config").Parse(nginxConfig)).Execute(buf, struct{ SocketPath, CurrentPath string }{"/tmp/rack.socket", "/app"})
	if e != nil {
		panic(e.Error())
	}
	return buf.String()
}

const upstart = `env PATH={{ .RubyInstallPath }}/bin:$PATH
chdir /app
exec puma -e production -b unix://{{ .SocketPath }} --pidfile /var/run/app.pid
`

const runSh = `
#!/bin/bash

export PATH={{ .RubyInstallPath }}/bin:$PATH
which puma || gem install puma --no-ri --no-rdoc
`

const configRu = `
app = lambda do |env|
  [200, { "Content-Type" => "text/plain" }, ["Hello from urknall!"]]
end

run app
`

const nginxConfig = `
syslog user nginx;

events {
  worker_connections  1024;
}

worker_processes 4;
pid /var/run/nginx.pid;

http {
  include mime.types;
  upstream rack {
      server unix:{{ .SocketPath }} fail_timeout=0;
  }

  log_format default
    'ip=$remote_addr forwarded=$http_x_forwarded_for host=$http_host method=$request_method status=$status length=$body_bytes_sent '
    'total=$request_time upstream_time=$upstream_response_time ua="$http_user_agent" uri="$request_uri" ref="$http_referer"';

  access_log syslog:notice default;
  error_log syslog:error;

  server {
    more_set_headers "Server: nginx";

    server_name _;
    listen 80;

    root {{ .CurrentPath }}/public;

    try_files $uri @rack;

    location @rack {
      proxy_set_header X-Forwarded-Proto $scheme;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header Host $http_host;
      proxy_redirect off;
      proxy_pass http://rack;
      proxy_temp_path /tmp/nginx;
    }
  }
}
`
