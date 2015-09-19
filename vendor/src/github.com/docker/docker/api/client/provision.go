package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/utils"
	"gopkg.in/yaml.v2"
)

type Provision struct {
	Machines map[string]Machine
}

type Options map[string]interface{}

type Machine struct {
	Driver    string
	Instances int
	Options   Options
}

func (o Options) String(key string) string {
	result := o[key]
	if s, ok := result.(string); ok {
		if s[0:1] == "$" {
			env := os.Getenv(s[1:])
			if env == "" {
				return s
			}
			return env
		}
		return s
	}
	return ""
}

func (o Options) StringSlice(key string) []string {
	result := o[key]
	if s, ok := result.([]string); ok {
		return s
	} else if s, ok := result.(string); ok {
		return []string{s}
	}
	return []string{}
}

func (o Options) Int(key string) int {
	result := o[key]
	if i, ok := result.(int); ok {
		return i
	} else if s, ok := result.(string); ok {
		i, _ := strconv.Atoi(s)
		return i
	}
	return 0
}

func (o Options) Bool(key string) bool {
	result := o[key]
	if b, ok := result.(bool); ok {
		return b
	} else if s, ok := result.(string); ok {
		return s == "true" || s == "yes"
	}
	return false
}

func readFile(file string) ([]byte, error) {
	result := []byte{}
	// file := "docker-provision.yml"

	log.Debugf("Opening provision file: %s", file)

	if file == "-" {
		if bytes, err := ioutil.ReadAll(os.Stdin); err != nil {
			log.Errorf("Failed to read compose file from stdin: %v", err)
			return nil, err
		} else {
			result = bytes
		}
	} else if file != "" {
		if bytes, err := ioutil.ReadFile(file); os.IsNotExist(err) {
			log.Errorf("Failed to find %s", file)
			return nil, err
		} else if err != nil {
			log.Errorf("Failed to open %s", file)
			return nil, err
		} else {
			result = bytes
		}
	}

	return result, nil
}

func Generate(pattern string) []string {
	re, _ := regexp.Compile(`\[(.+):(.+)\]`)
	submatch := re.FindStringSubmatch(pattern)
	if submatch == nil {
		return []string{pattern}
	}

	from, err := strconv.Atoi(submatch[1])
	if err != nil {
		return []string{pattern}
	}
	to, err := strconv.Atoi(submatch[2])
	if err != nil {
		return []string{pattern}
	}

	template := re.ReplaceAllString(pattern, "%d")

	var result []string
	for val := from; val <= to; val++ {
		entry := fmt.Sprintf(template, val)
		result = append(result, entry)
	}

	return result
}

func setupCertificates(caCertPath, caKeyPath, clientCertPath, clientKeyPath string) error {
	org := utils.GetUsername()
	bits := 2048

	if _, err := os.Stat(utils.GetMachineCertDir()); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(utils.GetMachineCertDir(), 0700); err != nil {
				log.Fatalf("Error creating machine config dir: %s", err)
			}
		} else {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		log.Infof("Creating CA: %s", caCertPath)

		// check if the key path exists; if so, error
		if _, err := os.Stat(caKeyPath); err == nil {
			log.Fatalf("The CA key already exists.  Please remove it or specify a different key/cert.")
		}

		if err := utils.GenerateCACertificate(caCertPath, caKeyPath, org, bits); err != nil {
			log.Infof("Error generating CA certificate: %s", err)
		}
	}

	if _, err := os.Stat(clientCertPath); os.IsNotExist(err) {
		log.Infof("Creating client certificate: %s", clientCertPath)

		if _, err := os.Stat(utils.GetMachineCertDir()); err != nil {
			if os.IsNotExist(err) {
				if err := os.Mkdir(utils.GetMachineCertDir(), 0700); err != nil {
					log.Fatalf("Error creating machine client cert dir: %s", err)
				}
			} else {
				log.Fatal(err)
			}
		}

		// check if the key path exists; if so, error
		if _, err := os.Stat(clientKeyPath); err == nil {
			log.Fatalf("The client key already exists.  Please remove it or specify a different key/cert.")
		}

		if err := utils.GenerateCert([]string{""}, clientCertPath, clientKeyPath, caCertPath, caKeyPath, org, bits); err != nil {
			log.Fatalf("Error generating client certificate: %s", err)
		}
	}

	return nil
}

// Usage: gattai provision
func (cli *DockerCli) CmdProvision(args ...string) error {
	cmd := Cli.Subcmd("provision", []string{"pattern"}, "Machine patterns, e.g. machine-[1:10]", false)
	provisionFilename := cmd.String([]string{"f", "-file"}, "docker-provision.yml", "Name of the provision file")

	// TODO: EnvVar: "MACHINE_STORAGE_PATH"
	machineStoragePath := cmd.String([]string{"s", "-storge-path"}, utils.GetBaseDir(), "Configure Docker Machine's storage path")

	cmd.ParseFlags(args, true)

	ssh.SetDefaultClient(ssh.Native)

	bytes, err := readFile(*provisionFilename)
	var p Provision
	err = yaml.Unmarshal(bytes, &p)
	log.Infof("%s", p)

	// verify driver
	for group, details := range p.Machines {
		if details.Instances == 0 {
			details.Instances = 1
		} else if details.Instances < 0 {
			return fmt.Errorf("group %s has incorrect instance: %d", group, details.Instances)
		}
		found := false
		for _, driver := range drivers.GetDriverNames() {
			if driver == details.Driver {
				found = true
				break
			}
		}
		if found == false {
			return fmt.Errorf("group %s uses non-existed driver: %s", group, details.Driver)
		}
	}

	// extract pattern
	log.Info(args)

	machineList := []string{}

	// list all
	if len(cmd.Args()) == 0 {

		for group, details := range p.Machines {
			pattern := fmt.Sprintf("%s-[1:%d]", group, details.Instances)
			machineList = append(machineList, Generate(pattern)...)
		}

	} else {

		for _, arg := range cmd.Args() {
			// if it's a group name, use all instances of the group
			if details, exist := p.Machines[arg]; exist {
				pattern := fmt.Sprintf("%s-[1:%d]", arg, details.Instances)
				machineList = append(machineList, Generate(pattern)...)
			} else {
				// assume it's a pattern
				machineList = append(machineList, Generate(arg)...)
			}

			// TODO detect bad pattern and reject them
			// Correct pattern are:
			//  - ${group}-[m:n]
			//  - ${group}-n
		}

	}

	log.Info(machineList)

	if len(machineList) == 0 {
		// return
	}

	// create libmachine's store
	log.Info(*machineStoragePath)

	caCertPath := filepath.Join(utils.GetMachineCertDir(), "ca.pem")
	caKeyPath := filepath.Join(utils.GetMachineCertDir(), "ca-key.pem")
	clientCertPath := filepath.Join(utils.GetMachineCertDir(), "cert.pem")
	clientKeyPath := filepath.Join(utils.GetMachineCertDir(), "key.pem")

	certInfo := libmachine.CertPathInfo{
		CaCertPath:     caCertPath,
		CaKeyPath:      caKeyPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
	}

	if err := setupCertificates(
		certInfo.CaCertPath,
		certInfo.CaKeyPath,
		certInfo.ClientCertPath,
		certInfo.ClientKeyPath); err != nil {
		log.Fatalf("Error generating certificates: %s", err)
	}

	store := libmachine.NewFilestore(
		*machineStoragePath,
		certInfo.CaCertPath,
		certInfo.CaKeyPath)

	provider, err := libmachine.New(store)

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
						TlsVerify:  true,
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

		url, _ := host.GetURL()
		fmt.Printf("%-20s%s\n", machineName, url)

		// then, check status to be active
		active, _ := host.IsActive()
		if active == false {
			// if not set active
			host.Start()
			// check info
		}
	}

	// loop {}
	//if everything is OK, fmt.Printf("checking %s ...\t ready\n", machine)

	return err
}
