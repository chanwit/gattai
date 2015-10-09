package client

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	"github.com/docker/machine/libmachine/swarm"
	"github.com/mattn/go-shellwords"
	// "github.com/pkg/sftp"
)

func removeAllContainers(h *host.Host) error {
	url, err := h.GetURL()
	if err != nil {
		return err
	}

	psArgs := append([]string{
		"-H", url,
		"--tlscacert=" + h.HostOptions.AuthOptions.CaCertPath,
		"--tlscert=" + h.HostOptions.AuthOptions.ClientCertPath,
		"--tlskey=" + h.HostOptions.AuthOptions.ClientKeyPath,
		"--tlsverify=true"},
		"ps", "-aq")
	ps := exec.Command(os.Args[0], psArgs...)
	bytes, err := ps.Output()
	if err != nil {
		return err
	}
	containers := strings.Split(strings.TrimSpace(string(bytes)), "\n")

	args := append([]string{
		"-H", url,
		"--tlscacert=" + h.HostOptions.AuthOptions.CaCertPath,
		"--tlscert=" + h.HostOptions.AuthOptions.ClientCertPath,
		"--tlskey=" + h.HostOptions.AuthOptions.ClientKeyPath,
		"--tlsverify=true",
		"rm", "-f"},
		containers...)

	rm := exec.Command(os.Args[0], args...)
	bytes, err = rm.CombinedOutput()
	if err != nil {
		fmt.Print(string(bytes))
		return err
	}

	return nil
}

func configureNetwork(kvstore, firstMachine *host.Host, opt machine.Options) (machine.Options, error) {
	//   engine-label:
	//     - "com.docker.network.driver.overlay.bind_interface=eth0"
	//     - "com.docker.network.driver.overlay.neighbor_ip=${NODE_1_IP}"
	//   engine-opt:
	//     - "default-network overlay:multihost"
	//     - "kv-store consul:${KVSTORE_IP}:8500"

	labels := opt.StringSlice("engine-label")
	labels = append(labels, "com.docker.network.driver.overlay.bind_interface=eth0")
	if firstMachine != nil {
		ip, err := firstMachine.Driver.GetIP()
		if err != nil {
			return nil, err
		}
		labels = append(labels, "com.docker.network.driver.overlay.neighbor_ip="+ip)
	}

	opts := opt.StringSlice("engine-opt")
	// opts = append(opts, "default-network overlay:multihost")
	ip, err := kvstore.Driver.GetIP()
	if err != nil {
		return nil, err
	}
	opts = append(opts, "cluster-store consul://"+ip+":8500")

	opt["engine-label"] = labels
	opt["engine-opt"] = opts

	return opt, nil
}

func engineExecute(h *host.Host, line string) error {
	p := shellwords.NewParser()
	args, err := p.Parse(line)
	if err != nil {
		return err
	}

	url, err := h.GetURL()
	if err != nil {
		return err
	}

	args = append([]string{
		"-H", url,
		"--tlscacert=" + h.HostOptions.AuthOptions.CaCertPath,
		"--tlscert=" + h.HostOptions.AuthOptions.ClientCertPath,
		"--tlskey=" + h.HostOptions.AuthOptions.ClientKeyPath,
		"--tlsverify=true"}, args...)

	cmd := exec.Command(os.Args[0], args...)
	log.Debugf("[engineExecute] Executing: %s", cmd)

	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Print(string(b))
		return err
	}

	return nil
}

// Usage: gattai provision
func DoProvision(cli interface{}, args ...string) error {

	cmd := Cli.Subcmd("provision",
		[]string{"PATTERNS"},
		"Provision a set of machines. Patterns, e.g. machine-[1:10], are allowed.",
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

	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Do not list machines at the end of provisioning")

	cmd.ParseFlags(args, true)

	p, err := machine.ReadProvision(*provisionFilename)
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

	spacing := len(machineList) > 1

	// check each machine existing
	for _, name := range machineList {

		parts := strings.SplitN(name, "-", 2)
		group := parts[0]
		details := p.Machines[group]

		h, err := store.Load(name)
		if err != nil {
			if _, ok := err.(mcnerror.ErrHostDoesNotExist); ok {
				fmt.Printf("Machine '%s' not found, creating...\n", name)
				spacing = true

				c := machine.Options(make(map[string]interface{}))
				for k, v := range details.Options {
					c[k] = v
				}

				if details.Network != "" && details.Network != "none" {
					kvstoreName := details.NetworkKvstore
					if kvstoreName == "" {
						return errors.New("No kv-store specified")
					}
					kvstore, err := store.Load(kvstoreName)
					if err != nil {
						return err
					}
					firstMachineName := fmt.Sprintf("%s-%d", group, 1)
					var firstMachine *host.Host
					if firstMachineName != name {
						firstMachine, err = store.Load(firstMachineName)
						if err != nil {
							return err
						}
					}

					c, err = configureNetwork(kvstore, firstMachine, c)
					if err != nil {
						return err
					}
				}

				//spew.Dump(c)

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

				// spew.Dump(hostOptions)

				driver, err := driverfactory.NewDriver(details.Driver, name, *machineStoragePath)
				if err != nil {
					log.Fatalf("Error trying to get driver: %s", err)
				}

				// TODO populate Env Vars from all hosts
				// to use with .SetConfigFromFlags

				h, err = store.NewHost(driver)
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
			fmt.Printf("Machine '%s' exists, starting...\n", name)
			h.Start()
			spacing = false
		}

		_ = removeAllContainers(h)
		// TODO delete all containers during re-provision?

		if details.PostProvision != nil && len(details.PostProvision) > 0 {
			fmt.Println("Processing post-provision commands...")
			for _, post := range details.PostProvision {
				log.Debugf("post-provision: %s", post)
				if strings.HasPrefix(post, "docker") {
					err := engineExecute(h, strings.TrimSpace(post[6:]))
					if err != nil {
						return err
					}
				}
			}
		}

		if spacing {
			if len(machineList) > 1 {
				fmt.Println()
			}
		}

	}

	if !spacing {
		fmt.Println()
	}

	// TODO
	// post-provision state checks (commands:)
	// for _, machineName := range machineList {
	//	host, err := provider.Get(machineName)
	//	// host.
	// }

	if *quiet == false {
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
	}

	return err
}
