package client

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/chanwit/gattai/machine"
	"github.com/chanwit/gattai/utils"
	Cli "github.com/docker/docker/cli"
)

func DoRmm(cli interface{}, args ...string) error {

	cmd := Cli.Subcmd("rmm", []string{"MACHINES"}, "Remove machines", false)

	force := cmd.Bool([]string{"f", "-force"}, false, "Force removing machines")

	cmd.ParseFlags(args, true)

	if len(cmd.Args()) == 0 {
		return fmt.Errorf("You must specify a machine or pattern name")
	}

	isError := false

	store := machine.GetDefaultStore(utils.GetBaseDir())

	for _, pattern := range cmd.Args() {
		for _, hostName := range utils.Generate(pattern) {
			h, err := loadHost(store, hostName, utils.GetBaseDir())
			if err != nil {
				isError = true
				log.Errorf("Error removing machine %s: %s", hostName, err)
			}

			if err := h.Driver.Remove(); err != nil {
				if !*force {
					isError = true
					log.Errorf("Provider error removing machine %q: %s", hostName, err)
					continue
				}
			}
			if err := store.Remove(hostName); err != nil {
				isError = true
				log.Errorf("Error removing machine %q from store: %s", hostName, err)
			} else {
				log.Infof("Successfully removed %s", hostName)
			}
		}
	}

	if isError {
		return fmt.Errorf("There was an error removing a machine.")
	}

	return nil
}
