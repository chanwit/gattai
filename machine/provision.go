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
	Driver    string
	Instances int
	Options   Options
	Commands  []Command
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

	return parseProvision(bytes)
}

func (p *Provision) VerifyDrivers() error {
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

	return nil
}
