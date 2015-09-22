package machine

import (
	"fmt"

	"github.com/chanwit/gattai/utils"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"gopkg.in/yaml.v2"
)

type Provision struct {
	Machines map[string]Machine
}

type Machine struct {
	Driver    string
	Instances int
	Options   Options
}

func ReadProvision(file string) (Provision, error) {
	bytes, err := utils.ReadFile(file)
	if err != nil {
		return Provision{}, err
	}

	var p Provision
	err = yaml.Unmarshal(bytes, &p)
	if err != nil {
		return Provision{}, err
	}

	log.Debugf("%s", p)
	return p, nil
}

func (p Provision) VerifyDrivers() error {
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
