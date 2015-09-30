package client

/*
import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/chanwit/gattai/machine"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/machine/libmachine/mcnutils"
)

func (cli *DockerCli) CmdRmm(args ...string) error {
	cmd := Cli.Subcmd("rmm", []string{"machines"}, "Remove machines", false)

	force := cmd.Bool([]string{"f", "-force"}, false, "Force removing machines")

	cmd.ParseFlags(args, true)

	if len(cmd.Args()) == 0 {
		return fmt.Errorf("You must specify a machine name")
	}

	isError := false

	certInfo := machine.GetCertInfo()
	provider, err := machine.GetProvider(utils.GetBaseDir(), certInfo)
	if err != nil {
		return err
	}

	for _, host := range cmd.Args() {
		if err := provider.Remove(host, *force); err != nil {
			log.Errorf("Error removing machine %s: %s", host, err)
			isError = true
		} else {
			log.Infof("Successfully removed %s", host)
		}
	}

	if isError {
		return fmt.Errorf("There was an error removing a machine. To force remove it, pass the -f option. Warning: this might leave it running on the provider.")
	}

	return nil
}
*/