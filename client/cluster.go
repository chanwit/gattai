package client

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/chanwit/gattai/machine"
	"github.com/chanwit/gattai/utils"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
)

func setRemoteAuthOptions(p provision.Provisioner) auth.AuthOptions {
	dockerDir := p.GetDockerOptionsDir()
	authOptions := p.GetAuthOptions()

	// due to windows clients, we cannot use filepath.Join as the paths
	// will be mucked on the linux hosts
	authOptions.CaCertRemotePath = path.Join(dockerDir, "ca.pem")
	authOptions.ServerCertRemotePath = path.Join(dockerDir, "server.pem")
	authOptions.ServerKeyRemotePath = path.Join(dockerDir, "server-key.pem")

	return authOptions
}

func swarmManage(h *host.Host, image string, token string) error {

	provisioner, err := provision.DetectProvisioner(h.Driver)
	dockerDir := provisioner.GetDockerOptionsDir()
	authOptions := setRemoteAuthOptions(provisioner)

	url, err := h.GetURL()
	if err != nil {
		return err
	}

	exec.Command(os.Args[0], []string{
		"-H", url,
		"--tlscacert=" + h.HostOptions.AuthOptions.CaCertPath,
		"--tlscert=" + h.HostOptions.AuthOptions.ClientCertPath,
		"--tlskey=" + h.HostOptions.AuthOptions.ClientKeyPath,
		"--tlsverify=true",
		"rm", "-f", "swarm-manager"}...).Output()

	cmd := exec.Command(os.Args[0], []string{
		"-H", url,
		"--tlscacert=" + h.HostOptions.AuthOptions.CaCertPath,
		"--tlscert=" + h.HostOptions.AuthOptions.ClientCertPath,
		"--tlskey=" + h.HostOptions.AuthOptions.ClientKeyPath,
		"--tlsverify=true",
		"run", "-d", "--restart=always",
		"--net=bridge",
		"--name", "swarm-manager",
		"-p", "3376:3376",
		"-v", dockerDir + ":" + dockerDir,
		image, "manage",
		"--tlsverify",
		"--tlscacert=" + authOptions.CaCertRemotePath,
		"--tlscert=" + authOptions.ServerCertRemotePath,
		"--tlskey=" + authOptions.ServerKeyRemotePath,
		"-H", "0.0.0.0:3376",
		"token://" + token}...)

	err = cmd.Run()
	if err == nil {
		fmt.Printf("Manager '%s' started successfully...\n", h.Name)
	}

	return err
}

func swarmJoin(name string, image string, token string) error {

	store := machine.GetDefaultStore(utils.GetBaseDir())
	h, err := store.Load(name)
	if err != nil {
		return err
	}

	url, err := h.GetURL()
	if err != nil {
		return err
	}

	ip, err := h.Driver.GetIP()
	if err != nil {
		return err
	}

	exec.Command(os.Args[0], []string{
		"-H", url,
		"--tlscacert=" + h.HostOptions.AuthOptions.CaCertPath,
		"--tlscert=" + h.HostOptions.AuthOptions.ClientCertPath,
		"--tlskey=" + h.HostOptions.AuthOptions.ClientKeyPath,
		"--tlsverify=true",
		"rm", "-f", "swarm-agent"}...).Output()

	cmd := exec.Command(os.Args[0], []string{
		"-H", url,
		"--tlscacert=" + h.HostOptions.AuthOptions.CaCertPath,
		"--tlscert=" + h.HostOptions.AuthOptions.ClientCertPath,
		"--tlskey=" + h.HostOptions.AuthOptions.ClientKeyPath,
		"--tlsverify=true",
		"run", "-d", "--restart=always",
		"--net=bridge",
		"--name", "swarm-agent",
		image, "join",
		"--advertise", ip + ":2376",
		"token://" + token}...)

	b, err := cmd.CombinedOutput()
	if err == nil {
		fmt.Printf("Machine '%s' joined cluster...\n", name)
	} else {
		fmt.Println(string(b))
	}

	return err
}

func DoCluster(cli interface{}, args ...string) error {
	cmd := Cli.Subcmd("cluster", []string{"MACHINES"},
		"Form a cluster with a set of specified machines", false)

	master := cmd.String([]string{"m", "-master"}, "", "Configure the cluster masters")
	image := cmd.String([]string{"i", "-image"}, "swarm", "Specify Docker Swarm image")
	provision := cmd.Bool([]string{"p", "-provision"}, false, "Automatic provision before forming the cluster")

	cmd.ParseFlags(args, true)

	if *master == "" && len(cmd.Args()) == 0 {
		return errors.New("Please specify a set of machines or the cluster master.")
	}

	// do provision if required
	if *provision == true {
		err := DoProvision(cli, append([]string{"-q", *master}, cmd.Args()...)...)
		if err != nil {
			return err
		}

		fmt.Println()
	}

	// Read existing token
	// If not existed, generate one
	// TODO support other discoveries
	token, err := readToken()
	if err != nil {
		token, err = generateToken()
		if err != nil {
			return err
		}
	}

	p, err := machine.ReadProvision("provision.yml")
	if err != nil {
		log.Debugf("err: %s", err)
		return err
	}

	fmt.Printf("Use discovery token://%s\n", token)

	// start
	for _, machineName := range p.GetMachineList(cmd.Args()...) {
		err := swarmJoin(machineName, *image, token)
		if err != nil {
			return err
		}
	}

	if master != nil {
		store := machine.GetDefaultStore(utils.GetBaseDir())
		h, err := store.Load(*master)
		if err != nil {
			return err
		}

		err = swarmManage(h, *image, token)
		if err != nil {
			return err
		}

		f, err := os.Create(ACTIVE_HOST_FILE)
		defer f.Close()
		if err != nil {
			return err
		}

		// save config file for cluster
		ip, _ := h.Driver.GetIP()
		fmt.Fprintf(f, "---\n")
		fmt.Fprintf(f, "name: %s\n", h.Name)
		fmt.Fprintf(f, "DOCKER_HOST: \"tcp://%s:%d\"\n", ip, 3376)
		fmt.Fprintf(f, "DOCKER_CERT_PATH: %s\n", h.HostOptions.AuthOptions.StorePath)
		fmt.Fprintf(f, "DOCKER_TLS_VERIFY: 1\n")
		fmt.Printf("Active host is now set to '%s' (swarm).\n", h.Name)
	}

	return nil
}
