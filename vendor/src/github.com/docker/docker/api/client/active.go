package client

import (
	"fmt"
	"os"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/utils"
	"gopkg.in/yaml.v2"
)

func (cli *DockerCli) CmdActive(args ...string) error {
	cmd := Cli.Subcmd("active", []string{"machine name"}, "Machine's name", false)
	provisionFilename := cmd.String(
		[]string{"f", "-file"},
		"provision.yml",
		"Name of the provision file")

	// TODO: EnvVar: "MACHINE_STORAGE_PATH"
	machineStoragePath := cmd.String(
		[]string{"s", "-storge-path"},
		utils.GetBaseDir(),
		"Configure Docker Machine's storage path")

	cmd.ParseFlags(args, true)

	ssh.SetDefaultClient(ssh.Native)

	_, err := readProvision(*provisionFilename)

	certInfo := GetCertInfo()

	provider, err := GetProvider(*machineStoragePath, certInfo)

	if len(args) == 0 {
		envs := make(map[string]string)
		bytes, err := readFile(".gattai/.active_host")
		err = yaml.Unmarshal(bytes, &envs)
		if err == nil {
			fmt.Println(envs["name"])
			return nil
		}
	}

	host, err := provider.Get(args[0])
	if err == nil {
 		f, err := os.Create(".gattai/.active_host")
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
	}

	return err
}
