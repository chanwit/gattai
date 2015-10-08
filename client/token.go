package client

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/chanwit/gattai/utils"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/swarm/discovery/token"
	"gopkg.in/yaml.v2"
)

func generateToken() (string, error) {
	// Create a new token
	f, err := os.Create(".gattai/.token")
	defer f.Close()
	if err != nil {
		return "", err
	}

	discovery := &token.Discovery{}
	discovery.Initialize("", 0, 0)
	tk, err := discovery.CreateCluster()
	if err != nil {
		return "", err
	}

	fmt.Fprintf(f, "---\n")
	fmt.Fprintf(f, "CLUSTER_TOKEN: %s\n", tk)

	return tk, nil
}

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

func deleteToken(tk string) error {
	c := &http.Client{}
	req, err := http.NewRequest("DELETE", "https://discovery.hub.docker.com/v1/clusters/"+tk, nil)
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

	return nil
}

func DoToken(cli interface{}, args ...string) error {

	cmd := Cli.Subcmd("token", []string{}, "Create token for a cluster on Docker Hub", false)

	delete := cmd.Bool([]string{"d", "-delete"}, false, "Delete token")

	cmd.ParseFlags(args, true)

	if len(cmd.Args()) != 0 {
		return errors.New("Invalid argument number")
	}

	if *delete {
		tk, err := readToken()
		if err != nil {
			return err
		}

		err = deleteToken(tk)
		if err != nil {
			return err
		}

		fmt.Println("Unset the token.")
		return nil
	}

	// Show the current token
	if tk, err := readToken(); err == nil {
		fmt.Printf("Token already exists: %s\n", tk)
		return nil
	}

	if tk, err := generateToken(); err == nil {
		fmt.Println(tk)
		return nil
	}

	return errors.New("Error generating token")
}
