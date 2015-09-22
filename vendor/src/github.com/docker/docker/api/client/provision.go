package client

import (
	"fmt"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/chanwit/gattai/machine"
	Utils "github.com/chanwit/gattai/utils"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/utils"
)

// Usage: gattai provision
func (cli *DockerCli) CmdProvision(args ...string) error {
	cmd := Cli.Subcmd("provision",
		[]string{"pattern"},
		"Machine patterns, e.g. machine-[1:10]",
		false)

	provisionFilename := cmd.String(
		[]string{"f", "-file"},
		"provision.yml",
		"Name of the provision file")

	// TODO: EnvVar: "MACHINE_STORAGE_PATH"
	machineStoragePath := cmd.String(
		[]string{"s", "-storge-path"},
		utils.GetBaseDir(),
		"Configure Docker Machine's storage path")

	cmd.ParseFlags(args, true)

	ssh.SetDefaultClient(ssh.Native)

	p, err := machine.ReadProvision(*provisionFilename)
	if err != nil {
		return err
	}

	// TODO verify .gattai
	// if not, return err

	err = p.VerifyDrivers()
	if err != nil {
		return err
	}

	// extract pattern
	log.Debug(args)

	machineList := []string{}

	// list all
	if len(cmd.Args()) == 0 {

		for group, details := range p.Machines {
			pattern := fmt.Sprintf("%s-[1:%d]", group, details.Instances)
			machineList = append(machineList, Utils.Generate(pattern)...)
		}

	} else {

		for _, arg := range cmd.Args() {
			// if it's a group name, use all instances of the group
			if details, exist := p.Machines[arg]; exist {
				pattern := fmt.Sprintf("%s-[1:%d]", arg, details.Instances)
				machineList = append(machineList, Utils.Generate(pattern)...)
			} else {
				// assume it's a pattern
				machineList = append(machineList, Utils.Generate(arg)...)
			}

			// TODO detect bad pattern and reject them
			// Correct pattern are:
			//  - ${group}-[m:n]
			//  - ${group}-n
		}

	}

	log.Debug(machineList)

	if len(machineList) == 0 {
		// return
	}

	// create libmachine's store
	log.Debug(*machineStoragePath)

	certInfo := machine.GetCertInfo()

	if err := machine.SetupCertificates(
		certInfo.CaCertPath,
		certInfo.CaKeyPath,
		certInfo.ClientCertPath,
		certInfo.ClientKeyPath); err != nil {
		log.Fatalf("Error generating certificates: %s", err)
	}

	provider, err := machine.GetProvider(*machineStoragePath, certInfo)

	// check each machine existing
	for _, machineName := range machineList {

		host, err := provider.Get(machineName)
		if err != nil {
			if _, ok := err.(libmachine.ErrHostDoesNotExist); ok {
				fmt.Printf("Machine '%s' not found, creating ...\n", machineName)
				parts := strings.SplitN(machineName, "-", 2)
				group := parts[0]
				details := p.Machines[group]
				// REF: docker/machine/commands/create.go#76
				hostOptions := &libmachine.HostOptions{
					AuthOptions: &auth.AuthOptions{
						CaCertPath:     certInfo.CaCertPath,
						PrivateKeyPath: certInfo.CaKeyPath,
						ClientCertPath: certInfo.ClientCertPath,
						ClientKeyPath:  certInfo.ClientKeyPath,
						ServerCertPath: filepath.Join(utils.GetMachineDir(), machineName, "server.pem"),
						ServerKeyPath:  filepath.Join(utils.GetMachineDir(), machineName, "server-key.pem"),
					},
					EngineOptions: &engine.EngineOptions{
						// TODO
						// ArbitraryFlags:   c.StringSlice("engine-opt"),
						// Env:              c.StringSlice("engine-env"),
						// InsecureRegistry: c.StringSlice("engine-insecure-registry"),
						// Labels:           c.StringSlice("engine-label"),
						// RegistryMirror:   c.StringSlice("engine-registry-mirror"),
						// StorageDriver:    c.String("engine-storage-driver"),
						TlsVerify: true,
						// TODO default values
						InstallURL: details.Options.String("engine-install-url"),
					},
					SwarmOptions: &swarm.SwarmOptions{},
				}
				host, err = provider.Create(machineName, details.Driver, hostOptions, details.Options)
				if err != nil {
					log.Errorf("Error creating machine: %s", err)
					log.Fatal("You will want to check the provider to make sure the machine and associated resources were properly removed.")
				}

			}
		}

		// url, _ := host.GetURL()
		// fmt.Printf("\t%-20s%s\n", machineName, url)

		// then, check status to be active
		active, _ := host.IsActive()
		if active == false {
			// if not set active
			host.Start()
			// check info
		}
	}

	fmt.Printf("%-16s%-30s%s\n", "NAME", "URL", "STATUS")
	for _, machineName := range machineList {
		host, err := provider.Get(machineName)
		if err == nil {
			url, _ := host.GetURL()
			fmt.Printf("%-16s%-30s%s\n", machineName, url, "ready")
		}
	}

	// loop {}
	//if everything is OK, fmt.Printf("checking %s ...\t ready\n", machine)

	return err
}
