package urknall

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/dynport/dgtk/pubsub"
	"github.com/dynport/gocli"
)

const (
	colorDryRun = 226
	colorCached = 33
	colorExec   = 34
)

var colorMapping = map[string]int{
	statusCached:       colorCached,
	statusExecFinished: colorExec,
}

var IgnoredMessagesError = errors.New("ignored published messages (subscriber buffer full)")

// Create a logging facility for urknall using urknall's default formatter.
// Note that this resource must be closed afterwards!
func OpenStdoutLogger() io.Closer {
	logger := &stdoutLogger{}
	logger.Formatter = logger.DefaultFormatter
	// Ignore the error from Start. It would only be triggered if the formatter wouldn't be set.
	_ = logger.Start()
	return logger
}

type stdoutLogger struct {
	Formatter    formatter
	maxLengths   map[int]int
	started      time.Time
	finished     chan interface{}
	pubSub       *pubsub.PubSub
	subscription *pubsub.Subscription
}

func (logger *stdoutLogger) Started() time.Time {
	if logger.started.IsZero() {
		logger.started = time.Now()
	}
	return logger.started
}

func (logger *stdoutLogger) formatCommandOuput(message *Message) string {
	prefix := fmt.Sprintf("[%s][%s][%s]", formatIp(message.HostIP()), formatRunlistName(message.RunlistName()), formatDuration(logger.sinceStarted()))
	line := message.line
	if message.IsStderr() {
		line = gocli.Red(line)
	}
	return prefix + " " + line
}

func formatIp(ip string) string {
	return fmt.Sprintf("%15s", ip)
}

type formatter func(urknallMessage *Message) string

func (logger *stdoutLogger) DefaultFormatter(message *Message) string {
	ignoreKeys := []string{MessageRunlistsPrecompile, MessageCleanupCacheEntries, MessageRunlistsProvision, MessageUrknallInternal}
	for _, k := range ignoreKeys {
		if strings.HasPrefix(message.Key(), k) {
			return ""
		}
	}
	if len(message.line) > 0 {
		return logger.formatCommandOuput(message)
	}
	ip := message.HostIP()
	runlistName := message.RunlistName()
	payload := ""
	if message.task != nil {
		payload = message.task.Command().Logging()
	}
	execStatus := fmt.Sprintf("%-8s", message.execStatus)
	if color := colorMapping[message.execStatus]; color > 0 {
		execStatus = gocli.Colorize(color, execStatus)
	}
	parts := []string{
		fmt.Sprintf("[%s][%s][%s][%s]%s",
			formatIp(ip),
			formatRunlistName(runlistName),
			formatDuration(logger.sinceStarted()),
			execStatus,
			payload,
		),
	}
	return strings.Join(parts, " ")
}

func formatRunlistName(name string) string {
	if len(name) > 8 {
		name = name[0:8]
	}
	return fmt.Sprintf("%-8s", name)
}

func formatDuration(dur time.Duration) string {
	durString := ""
	if dur >= 1*time.Millisecond {
		durString = fmt.Sprintf("%.03f", dur.Seconds())
	}
	return fmt.Sprintf("%7s", durString)
}

func (logger *stdoutLogger) sinceStarted() time.Duration {
	return time.Now().Sub(logger.Started())
}

func (logger *stdoutLogger) Start() error {
	logger.started = time.Now()
	if logger.Formatter == nil {
		return fmt.Errorf("Formatter must be set")
	}
	logger.pubSub = pubsub.New()
	RegisterPubSub(logger.pubSub)
	logger.subscription = logger.pubSub.Subscribe(func(m *Message) {
		if message := logger.Formatter(m); message != "" {
			log.Println(message)
		}
	})
	return nil
}

func (logger *stdoutLogger) Close() (e error) {
	e = logger.subscription.Close()
	if d := logger.pubSub.Stats.Ignored(); e == nil && d > 0 {
		return IgnoredMessagesError
	}
	return e
}

func init() {
	log.SetFlags(0)
}
