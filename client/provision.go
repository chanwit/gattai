package client

/*
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	log "github.com/Sirupsen/logrus"
	"github.com/chanwit/gattai/machine"
	Utils "github.com/chanwit/gattai/utils"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/pkg/sftp"
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
		log.Debugf("err: %s", err)
		return err
	}

	// TODO verify .gattai
	// if not, return err

	err = p.VerifyDrivers()
	if err != nil {
		log.Debugf("err: %s", err)
		return err
	}

	// extract pattern
	// fmt.Printf("args: %s\n",args)

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
				// if it's the only instance, use arg as name
				if details.Instances == 0 || details.Instances == 1 {
					machineList = append(machineList, arg)
				} else {
					pattern := fmt.Sprintf("%s-[1:%d]", arg, details.Instances)
					machineList = append(machineList, Utils.Generate(pattern)...)
				}
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

	log.Debugf("machines: %s", machineList)

	if len(machineList) == 0 {
		// return
	}

	// create libmachine's store
	log.Debugf("storage: %s", *machineStoragePath)

	certInfo := machine.GetCertInfo()

	// TODO authOptions :=

	if err := cert.BootstrapCertificates(authOptions); err != nil {
		log.Fatalf("Error generating certificates: %s", err)
	}

	store := machine.GetDefaultStore(*machineStoragePath)

	// check each machine existing
	for _, machineName := range machineList {

		host, err := store.Load(machineName) // provider.Get(machineName)
		if err != nil {
			if _, ok := err.(libmachine.ErrHostDoesNotExist); ok {
				fmt.Printf("Machine '%s' not found, creating ...\n", machineName)
				parts := strings.SplitN(machineName, "-", 2)
				group := parts[0]
				details := p.Machines[group]
				c := details.Options

				hostOptions := &host.HostOptions{
					AuthOptions: &auth.AuthOptions{
						CertDir:          mcndirs.GetMachineCertDir(),
						CaCertPath:       certInfo.CaCertPath,
						CaPrivateKeyPath: certInfo.CaPrivateKeyPath,
						ClientCertPath:   certInfo.ClientCertPath,
						ClientKeyPath:    certInfo.ClientKeyPath,
						ServerCertPath:   filepath.Join(mcndirs.GetMachineDir(), name, "server.pem"),
						ServerKeyPath:    filepath.Join(mcndirs.GetMachineDir(), name, "server-key.pem"),
						StorePath:        filepath.Join(mcndirs.GetMachineDir(), name),
					},
					EngineOptions: &engine.EngineOptions{
						ArbitraryFlags:   c.StringSlice("engine-opt"),
						Env:              c.StringSlice("engine-env"),
						InsecureRegistry: c.StringSlice("engine-insecure-registry"),
						Labels:           c.StringSlice("engine-label"),
						RegistryMirror:   c.StringSlice("engine-registry-mirror"),
						StorageDriver:    c.String("engine-storage-driver"),
						TlsVerify:        true,
						InstallURL:       c.String("engine-install-url"),
					},
					SwarmOptions: &swarm.SwarmOptions{
						IsSwarm:        c.Bool("swarm"),
						Image:          c.String("swarm-image"),
						Master:         c.Bool("swarm-master"),
						Discovery:      c.String("swarm-discovery"),
						Address:        c.String("swarm-addr"),
						Host:           c.String("swarm-host"),
						Strategy:       c.String("swarm-strategy"),
						ArbitraryFlags: c.StringSlice("swarm-opt"),
					},
				}

				driver, err := driverfactory.NewDriver(driverName, name, storePath)
				if err != nil {
					log.Fatalf("Error trying to get driver: %s", err)
				}

				h, err := store.NewHost(driver)
				if err != nil {
					log.Fatalf("Error getting new host: %s", err)
				}

				h.HostOptions = hostOptions

				/*
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
				}* /

				// host :=
				err := libmachine.Create(store, host)
				// host, err = provider.Create(machineName, details.Driver, hostOptions, details.Options)
				if err != nil {
					log.Errorf("Error creating machine: %s", err)
					log.Fatal("You will want to check the provider to make sure the machine and associated resources were properly removed.")
				}

			}
		}

		// then, check status to be active
		active, _ := host.IsActive()
		if active == false {
			// if not set active
			host.Start()
			// check info
		}
	}

	// post-provision state checks (commands:)
	// for _, machineName := range machineList {
	//	host, err := provider.Get(machineName)
	//	// host.
	// }

	w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tURL\tSTATE")

	for _, machineName := range machineList {
		host, err := provider.Get(machineName)
		items := libmachine.GetHostListItems([]*libmachine.Host{host})
		if err == nil {
			url, _ := host.GetURL()
			fmt.Fprintf(w, "%s\t%s\t%s\n", machineName, url, items[0].State)
		}
	}
	w.Flush()

	return err
}
*/
