package client

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/chanwit/gattai/machine"
	"github.com/chanwit/gattai/utils"
	Cli "github.com/docker/docker/cli"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/provision"
)

func startZooKeeper(args ...string) error {

	p, err := machine.ReadProvision("provision.yml")
	if err != nil {
		log.Debugf("err: %s", err)
		return err
	}

	// extract pattern
	// fmt.Printf("args: %s\n",args)

	machineList := p.GetMachineList(args...)

	log.Debugf("machines: %s", machineList)

	if len(machineList) == 0 {
		return errors.New("no machine in list")
	}

	store := machine.GetDefaultStore(utils.GetBaseDir())

	zooCfg := `
tickTime=2000
dataDir=/var/zookeeper/
clientPort=2181
initLimit=5
syncLimit=2
{{range $id, $ip := .}}server.{{ inc $id }}={{ $ip }}:2888:3888
{{ end }}
`
	addHosts := []string{}
	hosts := []*host.Host{}
	for _, name := range machineList {
		h, err := loadHost(store, name, utils.GetBaseDir())
		if err != nil {
			return err
		}

		hosts = append(hosts, h)
		ip, err := h.Driver.GetIP()
		if err != nil {
			return err
		}

		addHosts = append(addHosts, "--add-host", name+":"+ip)
	}

	tmpl, err := template.New("zoo").Funcs(template.FuncMap{
		"inc": func(i int) int { return i + 1 },
	}).Parse(zooCfg)

	if err != nil {
		return err
	}
	var str bytes.Buffer
	err = tmpl.Execute(&str, machineList) // change ips to machine name
	if err != nil {
		return err
	}

	for _, h := range hosts {

		provisioner, err := provision.DetectProvisioner(h.Driver)
		// dockerDir := provisioner.GetDockerOptionsDir()
		// authOptions := setRemoteAuthOptions(provisioner)

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
			"rm", "-f", "swarm-discovery-service"}...).Output()

		// skip error checking
		provisioner.SSHCommand("sudo mkdir -p /opt/zookeeper/conf")

		// copy zooCfg
		transferCmdFmt := "printf '%%s' '%s' | sudo tee %s"
		if _, err := provisioner.SSHCommand(fmt.Sprintf(transferCmdFmt, str.String(), "/opt/zookeeper/conf/zoo.cfg")); err != nil {
			return err
		}

		cmd := exec.Command(os.Args[0], append([]string{
			"-H", url,
			"--tlscacert=" + h.HostOptions.AuthOptions.CaCertPath,
			"--tlscert=" + h.HostOptions.AuthOptions.ClientCertPath,
			"--tlskey=" + h.HostOptions.AuthOptions.ClientKeyPath,
			"--tlsverify=true",
			"run", "-d", "--restart=always",
			"--name", "swarm-discovery-service",
			"-v", "/opt/zookeeper/conf:/opt/zookeeper/conf"},
			append(addHosts,
				"-p", "2181:2181",
				"-p", "2888:2888",
				"-p", "3888:3888",
				"jplock/zookeeper")...)...)

		output, err := cmd.CombinedOutput()
		if err == nil {
			fmt.Printf("ZK '%s' started successfully...\n", h.Name)
		}
		if err != nil {
			fmt.Print(string(output))
			return err
		}

	}

	return nil
}

// $ gattai discovery -t zk master
func DoDiscovery(cli interface{}, args ...string) error {
	cmd := Cli.Subcmd("discovery",
		[]string{"MACHINES"},
		"Set the machine specified as the discovery service", false)

	discoveryType := cmd.String(
		[]string{"t", "-discovery-type"}, "",
		"Configure a type of discovery service (consul, etcd, zk, token)")

	// path :
	_ = cmd.String([]string{"p", "-path"}, "/cluster", "Set discovery path")

	cmd.ParseFlags(args, true)

	if *discoveryType == "" {
		return errors.New("Please specify a type of discovery (consul, etcd, zk, token) ")
	}

	if *discoveryType == "token" && len(cmd.Args()) > 0 {
		return errors.New("Discovery 'token' does not require machine names")
	}

	switch *discoveryType {
	case "zk":
		err := DoProvision(cli, append([]string{"-q"}, cmd.Args()...)...)
		if err != nil {
			return err
		}
		err = startZooKeeper(cmd.Args()...)
		if err != nil {
			return err
		}
	}

	return nil
}
