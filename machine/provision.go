package machine

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/chanwit/gattai/utils"
	"github.com/docker/machine/libmachine/drivers"
	"gopkg.in/yaml.v2"
)

type Provision struct {
	Machines map[string]Machine
}

type Machine struct {
	From           string
	Driver         string
	Instances      int
	Options        Options
	Commands       []Command
	Network        string   // default is "" == none
	NetworkKvstore string   `yaml:"cluster-store"`
	PostProvision  []string `yaml:"post-provision"`
}

type Command map[string]string

func (c Command) Parse() map[string]string {
	s := "" // string(c)
	_ = strings.Split(s, "")
	return nil
}

func parseProvision(bytes []byte) (*Provision, error) {
	var p Provision
	err := yaml.Unmarshal(bytes, &p)
	if err != nil {
		return nil, err
	}

	log.Debugf("%s", p)
	return &p, nil
}

func ReadProvision(file string) (*Provision, error) {
	bytes, err := utils.ReadFile(file)
	if err != nil {
		return nil, err
	}

	p, err := parseProvision(bytes)
	if err != nil {
		return nil, err
	}

	err = p.verifyDrivers()
	if err != nil {
		return nil, err
	}

	return p, nil
}

// TODO change to verify
func (p *Provision) verifyDrivers() error {

	for group, details := range p.Machines {
		if details.Instances == 0 {
			details.Instances = 1
		} else if details.Instances < 0 {
			return fmt.Errorf("group %s has incorrect instance: %d", group, details.Instances)
		}

		// inherit drive and details with From
		if details.From != "" {
			from := p.Machines[details.From]
			details.Driver = from.Driver
			if details.Options == nil {
				details.Options = make(map[string]interface{})
			}
			for k, v := range from.Options {
				details.Options[k] = v
			}
			copy(details.Commands, from.Commands)
			p.Machines[group] = details
		}

		// verify driver
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

	return nil
}

func preSplit(patterns ...string) []string {
	result := []string{}
	for _, p := range patterns {
		result = append(result, strings.Split(p, ",")...)
	}
	return result
}

func (p *Provision) GetMachineList(patterns ...string) []string {

	patterns = preSplit(patterns...)

	machineList := []string{}

	// if patterns is blank, get all
	args := []string{}
	if len(patterns) == 0 {
		for group, _ := range p.Machines {
			args = append(args, group)
		}
	} else {
		args = append(args, patterns...)
	}

	for _, arg := range args {

		if details, exist := p.Machines[arg]; exist {
			// if it's the only instance, use arg as name
			if details.Instances == 0 || details.Instances == 1 {
				machineList = append(machineList, arg)
			} else {
				pattern := fmt.Sprintf("%s-[1:%d]", arg, details.Instances)
				machineList = append(machineList, utils.Generate(pattern)...)
			}

		} else {
			// assume it's a pattern
			machineList = append(machineList, utils.Generate(arg)...)
		}

	}

	return machineList
}
