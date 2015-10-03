package client

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	log "github.com/Sirupsen/logrus"
	"github.com/chanwit/gattai/machine"
	"github.com/chanwit/gattai/utils"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/drivers/driverfactory"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/swarm"
	// "github.com/pkg/sftp"
)

// Usage: gattai provision
func DoProvision(cli interface{}, args ...string) error {

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

	machineList := p.GetMachineList(cmd.Args()...)

	log.Debugf("machines: %s", machineList)

	if len(machineList) == 0 {
		return errors.New("no machine in list")
	}

	// create libmachine's store
	log.Debugf("storage: %s", *machineStoragePath)

	certInfo := machine.GetCertInfo()
	authOptions := &auth.AuthOptions{
		CertDir:          filepath.Join(*machineStoragePath, "certs"),
		CaCertPath:       certInfo.CaCertPath,
		CaPrivateKeyPath: certInfo.CaPrivateKeyPath,
		ClientCertPath:   certInfo.ClientCertPath,
		ClientKeyPath:    certInfo.ClientKeyPath,
	}

	// TODO authOptions :=

	if err := cert.BootstrapCertificates(authOptions); err != nil {
		log.Fatalf("Error generating certificates: %s", err)
	}

	store := machine.GetDefaultStore(*machineStoragePath)

	// check each machine existing
	for _, name := range machineList {

		h, err := store.Load(name)
		if err != nil {
			if _, ok := err.(mcnerror.ErrHostDoesNotExist); ok {
				fmt.Printf("Machine '%s' not found, creating...\n", name)
				parts := strings.SplitN(name, "-", 2)
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

				driver, err := driverfactory.NewDriver(details.Driver, name, *machineStoragePath)
				if err != nil {
					log.Fatalf("Error trying to get driver: %s", err)
				}

				h, err := store.NewHost(driver)
				if err != nil {
					log.Fatalf("Error getting new host: %s", err)
				}

				h.HostOptions = hostOptions

				if err := h.Driver.SetConfigFromFlags(details.Options); err != nil {
					log.Fatalf("Error setting machine configuration from flags provided: %s", err)
				}

				err = libmachine.Create(store, h)
				if err != nil {
					log.Errorf("Error creating machine: %s", err)
					log.Fatal("You will want to check the provider to make sure the machine and associated resources were properly removed.")
				}

			}
		} else {
			fmt.Printf("Machine '%s' existed, starting...\n", name)
			h.Start()
		}

		fmt.Println()
	}

	// TODO
	// post-provision state checks (commands:)
	// for _, machineName := range machineList {
	//	host, err := provider.Get(machineName)
	//	// host.
	// }

	w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tURL\tSTATE")

	for _, machineName := range machineList {
		h, err := store.Load(machineName)
		items := getHostListItems([]*host.Host{h})
		if err == nil {
			url, _ := h.GetURL()
			fmt.Fprintf(w, "%s\t%s\t%s\n", machineName, url, items[0].State)
		}
	}
	w.Flush()

	return err
}
