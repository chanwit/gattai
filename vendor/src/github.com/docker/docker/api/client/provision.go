package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	log "github.com/Sirupsen/logrus"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/utils"
	"gopkg.in/yaml.v2"
)

type Provision struct {
	Machines map[string]Machine
}

type Machine struct {
	Driver    string
	Instances int
	Options   []string
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

// Usage: gattai provision
func (cli *DockerCli) CmdProvision(args ...string) error {
	cmd := Cli.Subcmd("provision", []string{"pattern"}, "Machine patterns, e.g. machine-[1:10]", false)
	provisionFilename := cmd.String([]string{"f", "-file"}, "docker-provision.yml", "Name of the provision file")

	// TODO: EnvVar: "MACHINE_STORAGE_PATH"
	machineStoragePath := cmd.String([]string{"s", "-storge-path"}, utils.GetBaseDir(), "Configure Docker Machine's storage path")

	cmd.ParseFlags(args, true)

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

	store := libmachine.NewFilestore(
		*machineStoragePath,
		certInfo.CaCertPath,
		certInfo.CaKeyPath)

	provider, err := libmachine.New(store)

	// check each machine existing
	for _, machine := range machineList {

		//fmt.Printf("checking %s ...\t ready\n", machine)
		_, err := provider.Get(machine)
		if err != nil {
			if _, ok := err.(libmachine.ErrHostDoesNotExist); ok {
				fmt.Printf("machine %s not found\n", machine)
				// TODO creating
			}
		}
		// if not do provision

		// if so, check active

		// if not set active
	}
	return err
}
