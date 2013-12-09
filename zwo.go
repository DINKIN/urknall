// zwo provides everything necessary to provision machines, i.e. the  mechanisms required to run a set of commands
// somewhere (wherever that is, some bare metal or a docker container). These commands can be encapsulated in a package
// that has some configuration so that reuse is possible. There is even annotation based validation for packages.
//
// Every package compiles its raw commands into a runlist that is precompiled (find errors prior to running the first
// remote command), has variable substitution for some commands (the package's fields can be used in commands using go's
// templating mechanisms), and run on the respective host. This allows for provisioning packages in different
// configurations on different hosts.
//
// For each package a caching mechanism is used, so that repeated provisioning of the same package will only run the
// commands necessary (a changed command and all subsequent ones). This allows for extension and modification of the
// underlying host and takes the burden of writing idempotent commands from the developer. But in most cases it's more
// favorable to have throw away resources, that can easily replaced by a fresh one provisioned from ground up.
package zwo
