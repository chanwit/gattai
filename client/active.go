package client

import (
	"errors"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/chanwit/gattai/machine"
	"github.com/chanwit/gattai/utils"
	Cli "github.com/docker/docker/cli"
	"gopkg.in/yaml.v2"
)

const ACTIVE_HOST_FILE = ".gattai/.active_host"

func GetActiveHostName() (string, error) {
	envs := make(map[string]string)
	bytes, err := utils.ReadFile(ACTIVE_HOST_FILE)
	if err != nil {
		return "", errors.New("There is no active host.")
	}

	err = yaml.Unmarshal(bytes, &envs)
	if err == nil {
		return envs["name"], nil
	}

	return "", err
}

func DoActive(cli interface{}, args ...string) error {

	cmd := Cli.Subcmd("active",
		[]string{"machine name"},
		"Set the machine specified as the active Docker host (-- to unset)", false)

	machineStoragePath := cmd.String(
		[]string{"s", "-storge-path"},
		utils.GetBaseDir(),
		"Configure Docker Machine's storage path")

	master := cmd.Bool([]string{"m", "-master"}, false, "Active host is a Docker Swarm manager?")

	insecure := cmd.Bool([]string{"i", "-insecure"}, false, "Set the host active without TLS verification over 2375")

	cmd.ParseFlags(args, true)

	if len(cmd.Args()) == 0 && len(args) >= 1 && args[0] == "--" {
		err := os.Remove(ACTIVE_HOST_FILE)
		if err == nil {
			fmt.Println("Unset the active host.")
		}
		return err
	}

	if len(cmd.Args()) == 0 {
		name, err := GetActiveHostName()
		if err == nil {
			fmt.Println(name)
		}
		return err
	}

	store := machine.GetDefaultStore(*machineStoragePath)

	host, err := store.Load(cmd.Args()[0])
	if err == nil {
		f, err := os.Create(ACTIVE_HOST_FILE)
		defer f.Close()
		if err != nil {
			return err
		}

		// save active config
		fmt.Fprintf(f, "---\n")
		fmt.Fprintf(f, "name: %s\n", host.Name)
		ip, _ := host.Driver.GetIP()
		if *master {
			fmt.Fprintf(f, "DOCKER_HOST: \"tcp://%s:3376\"\n", ip)
		} else {
			url, _ := host.GetURL()
			if *insecure {
				fmt.Fprintf(f, "DOCKER_HOST: \"tcp://%s:2375\"\n", ip)
			} else {
				fmt.Fprintf(f, "DOCKER_HOST: \"%s\"\n", url)
			}
		}

		if *insecure == false {
			fmt.Fprintf(f, "DOCKER_CERT_PATH: %s\n", host.HostOptions.AuthOptions.StorePath)
			fmt.Fprintf(f, "DOCKER_TLS_VERIFY: 1\n")
		} else {
			fmt.Fprintf(f, "DOCKER_TLS_VERIFY: 0\n")
		}

		fmt.Println(cmd.Args()[0])
	} else {
		log.Error(err)
	}

	// err = vc.Commit(ACTIVE_HOST_FILE, "update host file")

	return err
}
