package client

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/machine/utils"
	"gopkg.in/yaml.v2"
)

const ACTIVE_HOST_FILE = ".gattai/.active_host"

func (cli *DockerCli) CmdActive(args ...string) error {
	cmd := Cli.Subcmd("active",
		[]string{"machine name"},
		"Set the machine specified as the active Docker host (-- to unset)", false)

	// TODO: EnvVar: "MACHINE_STORAGE_PATH"
	machineStoragePath := cmd.String(
		[]string{"s", "-storge-path"},
		utils.GetBaseDir(),
		"Configure Docker Machine's storage path")

	cmd.ParseFlags(args, true)

	if len(args) == 0 {
		envs := make(map[string]string)
		bytes, err := readFile(ACTIVE_HOST_FILE)
		if err != nil {
			fmt.Println("There is no active host.")
			return nil
		}

		err = yaml.Unmarshal(bytes, &envs)
		if err == nil {
			fmt.Println(envs["name"])
		}
		return err

	} else if len(args) == 1 && args[0] == "--" {
		err := os.Remove(ACTIVE_HOST_FILE)
		if err == nil {
			fmt.Println("Unset the active host.")
		}
		return err
	}

	// ssh.SetDefaultClient(ssh.Native)

	// _, err := readProvision(*provisionFilename)

	certInfo := GetCertInfo()

	provider, err := GetProvider(*machineStoragePath, certInfo)
	if err != nil {
		log.Error(err)
	}

	host, err := provider.Get(args[0])
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
		fmt.Fprintf(f, "DOCKER_CERT_PATH: %s\n", host.StorePath)
		fmt.Fprintf(f, "DOCKER_TLS_VERIFY: 1\n")

		fmt.Println(args[0])
	} else {
		log.Error(err)
	}

	return err
}
