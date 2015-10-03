package client

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/chanwit/gattai/machine"
	"github.com/chanwit/gattai/utils"
	Cli "github.com/docker/docker/cli"
	"gopkg.in/yaml.v2"
)

const ACTIVE_HOST_FILE = ".gattai/.active_host"

func DoActive(cli interface{}, args ...string) error {

	cmd := Cli.Subcmd("active",
		[]string{"machine name"},
		"Set the machine specified as the active Docker host (-- to unset)", false)

	machineStoragePath := cmd.String(
		[]string{"s", "-storge-path"},
		utils.GetBaseDir(),
		"Configure Docker Machine's storage path")

	cmd.ParseFlags(args, true)

	if len(cmd.Args()) == 0 {
		envs := make(map[string]string)
		bytes, err := utils.ReadFile(ACTIVE_HOST_FILE)
		if err != nil {
			fmt.Println("There is no active host.")
			return nil
		}

		err = yaml.Unmarshal(bytes, &envs)
		if err == nil {
			fmt.Println(envs["name"])
		}
		return err

	} else if len(cmd.Args()) == 1 && args[0] == "--" {
		err := os.Remove(ACTIVE_HOST_FILE)
		if err == nil {
			fmt.Println("Unset the active host.")
		}
		return err
	}

	// ssh.SetDefaultClient(ssh.Native)

	store := machine.GetDefaultStore(*machineStoragePath)

	host, err := store.Load(cmd.Args()[0])
	if err == nil {
		f, err := os.Create(ACTIVE_HOST_FILE)
		defer f.Close()
		if err != nil {
			return err
		}

		// save active config
		url, _ := host.GetURL()
		fmt.Fprintf(f, "---\n")
		fmt.Fprintf(f, "name: %s\n", host.Name)
		fmt.Fprintf(f, "DOCKER_HOST: \"%s\"\n", url)
		fmt.Fprintf(f, "DOCKER_CERT_PATH: %s\n", host.HostOptions.AuthOptions.StorePath)
		fmt.Fprintf(f, "DOCKER_TLS_VERIFY: 1\n")

		fmt.Println(cmd.Args()[0])
	} else {
		log.Error(err)
	}

	// err = vc.Commit(ACTIVE_HOST_FILE, "update host file")

	return err
}
