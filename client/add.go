package client

import (
	"errors"
	"io/ioutil"

	"github.com/chanwit/gattai/machine"
	Cli "github.com/docker/docker/cli"
	"gopkg.in/yaml.v2"
)

// Add a pre-defined flavor to provision file
// $ gattai add --flavor digitalocean-2g node
func DoAdd(cli interface{}, args ...string) error {
	cmd := Cli.Subcmd("add",
		[]string{"MACHINES"},
		"Add a pre-defined flavor",
		false)

	flavor := cmd.String([]string{"f", "-flavor"}, "", "Name of pre-defined flavor")
	n := cmd.Int([]string{"n", "--instances"}, 1, "Number of instances")

	cmd.ParseFlags(args, true)

	if *flavor == "" {
		return errors.New("Please specify a pre-set flavor")
	}

	p, err := machine.ReadRawProvision("provision.yml")
	if err != nil {
		return err
	}
	if p.Machines == nil {
		p.Machines = make(map[string]machine.Machine)
	}

	if len(cmd.Args()) != 1 {
		return errors.New("Please specify machine name")
	}

	machineGroup := cmd.Args()[0]

	switch *flavor {

	case "do-2g", "digitalocean-2g":
		p.Machines[machineGroup] = machine.Machine{
			Driver:    "digitalocean",
			Instances: *n,
			Options: map[string]interface{}{
				"digitalocean-image":        "ubuntu-14-04-x64",
				"digitalocean-region":       "nyc3",
				"digitalocean-size":         "2gb",
				"digitalocean-access-token": "$DIGITALOCEAN_ACCESS_TOKEN",
				"engine-install-url":        "https://get.docker.com",
			},
		}

	case "do-2g-ext", "digitalocean-2g-ext":
		p.Machines[machineGroup] = machine.Machine{
			Driver:    "digitalocean",
			Instances: *n,
			Options: map[string]interface{}{
				"digitalocean-image":        "debian-8-x64",
				"digitalocean-region":       "nyc3",
				"digitalocean-size":         "2gb",
				"digitalocean-access-token": "$DIGITALOCEAN_ACCESS_TOKEN",
				"engine-install-url":        "https://experimental.docker.com",
			},
		}

	case "do-2g-cluster", "digitalocean-2g-cluster":
		p.Machines[machineGroup+"-master"] = machine.Machine{
			Driver:    "digitalocean",
			Instances: 1,
			Options: map[string]interface{}{
				"digitalocean-image":        "debian-8-x64",
				"digitalocean-region":       "nyc3",
				"digitalocean-size":         "2gb",
				"digitalocean-access-token": "$DIGITALOCEAN_ACCESS_TOKEN",
				"engine-install-url":        "https://experimental.docker.com",
			},
			PostProvision: []string{
				"docker run -d -p 8400:8400 -p 8500:8500 -p 8600:53/udp progrium/consul --server -bootstrap-expect 1",
			},
		}

		p.Machines[machineGroup] = machine.Machine{
			Driver:         "digitalocean",
			Instances:      *n,
			NetworkKvstore: machineGroup + "-master",
			// Network: "overlay",
			Options: map[string]interface{}{
				"digitalocean-image":        "debian-8-x64",
				"digitalocean-region":       "nyc3",
				"digitalocean-size":         "2gb",
				"digitalocean-access-token": "$DIGITALOCEAN_ACCESS_TOKEN",
				"engine-install-url":        "https://experimental.docker.com",
			},
			PostProvision: []string{
				"docker network create -d overlay multihost",
			},
		}
	}

	provisionYml, err := yaml.Marshal(p)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("provision.yml", provisionYml, 0644)
	if err != nil {
		return err
	}

	return nil
}
