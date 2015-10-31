package client

import (
	"errors"
	"io/ioutil"

	"github.com/chanwit/gattai/flavor"
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

	flavorName := cmd.String([]string{"f", "-flavor"}, "", "Name of pre-defined flavor")
	n := cmd.Int([]string{"n", "--instances"}, 1, "Number of instances")

	cmd.ParseFlags(args, true)

	if *flavorName == "" {
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

	switch *flavorName {

	case "do-2g", "digitalocean-2g":
		f := flavor.DigitalOcean_2G
		f.Instances = *n
		p.Machines[machineGroup] = f

	case "do-2g-exp", "digitalocean-2g-exp":
		f := flavor.DigitalOcean_2G_Exp
		f.Instances = *n
		p.Machines[machineGroup] = f

	case "do-2g-cluster", "digitalocean-2g-cluster":
		master := flavor.DigitalOcean_2G_Cluster["master"]
		master.Instances = 1
		p.Machines[machineGroup+"-master"] = master

		node := flavor.DigitalOcean_2G_Cluster["node"]
		node.Instances = *n
		node.NetworkKvstore = machineGroup + "-master"
		p.Machines[machineGroup] = node
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
