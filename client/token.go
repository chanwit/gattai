package client

import (
	"fmt"
	"net/http"
	"os"

	"github.com/chanwit/gattai/utils"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/swarm/discovery/token"
	"gopkg.in/yaml.v2"
)

func readToken() (string, error) {
	envs := make(map[string]string)
	bytes, err := utils.ReadFile(".gattai/.token")
	if err != nil {
		return "", fmt.Errorf("There is no token set.")
	}

	err = yaml.Unmarshal(bytes, &envs)
	if err == nil {
		return envs["CLUSTER_TOKEN"], nil
	}

	return "", err
}

func DoToken(cli interface{}, args ...string) error {

	cmd := Cli.Subcmd("token", []string{}, "Create token for a cluster on Docker Hub", false)

	delete := cmd.Bool([]string{"d", "-delete"}, false, "Delete token")

	cmd.ParseFlags(args, true)

	if len(cmd.Args()) != 0 {
		return fmt.Errorf("Invalid argument number")
	}

	if *delete {
		token, err := readToken()
		if err != nil {
			return err
		}

		c := &http.Client{}
		req, err := http.NewRequest("DELETE", "https://discovery.hub.docker.com/v1/clusters/"+token, nil)
		if err != nil {
			return err
		}
		resp, err := c.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("Error code: %d", resp.StatusCode)
		}

		err = os.Remove(".gattai/.token")
		if err != nil {
			return err
		}

		fmt.Println("Unset the token.")
		return nil
	}

	// Show the current token
	tk, err := readToken()
	if err == nil {
		fmt.Printf("Token already existed: %s\n", tk)
		return nil
	}

	// Create a new token
	f, err := os.Create(".gattai/.token")
	defer f.Close()
	if err != nil {
		return err
	}

	discovery := &token.Discovery{}
	discovery.Initialize("", 0, 0)
	token, err := discovery.CreateCluster()
	if err != nil {
		return err
	}

	fmt.Fprintf(f, "---\n")
	fmt.Fprintf(f, "CLUSTER_TOKEN: %s\n", token)

	fmt.Println(token)
	return nil
}
