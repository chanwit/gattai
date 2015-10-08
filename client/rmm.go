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
		for _, host := range utils.Generate(pattern) {
			if err := store.Remove(host, *force); err != nil {
				log.Errorf("Error removing machine %s: %s", host, err)
				isError = true
			} else {
				log.Infof("Successfully removed %s", host)
			}
		}
	}

	if isError {
		return fmt.Errorf("There was an error removing a machine. To force remove it, pass the -f option. Warning: this might leave it running on the provider.")
	}

	return nil
}
