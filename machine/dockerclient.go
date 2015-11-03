package machine

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/docker/machine/libmachine/host"
	"github.com/samalba/dockerclient"
)

type Runner host.Host

func loadTLSConfig(ca, cert, key string, verify bool) (*tls.Config, error) {
	c, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load X509 key pair (%s, %s): %s. Key encrypted?",
			cert, key, err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{c},
		MinVersion:   tls.VersionTLS10,
	}

	if verify {
		certPool := x509.NewCertPool()
		file, err := ioutil.ReadFile(ca)
		if err != nil {
			return nil, fmt.Errorf("Couldn't read CA certificate: %s", err)
		}
		certPool.AppendCertsFromPEM(file)
		config.RootCAs = certPool
		config.ClientAuth = tls.RequireAndVerifyClientCert
		config.ClientCAs = certPool
	} else {
		// If --tlsverify is not supplied, disable CA validation.
		config.InsecureSkipVerify = true
	}

	return config, nil
}

func (h *Runner) ListAllContainers() ([]string, error) {
	cli, err := h.newClient()
	if err != nil {
		return []string{}, err
	}

	result := []string{}
	containers, err := cli.ListContainers(true, false, "")
	for _, c := range containers {
		result = append(result, c.Id)
	}

	return result, nil
}

func (h *Runner) RemoveAllContainers() error {
	cli, err := h.newClient()
	if err != nil {
		return err
	}

	containers, err := cli.ListContainers(true, false, "")
	for _, c := range containers {
		err := cli.RemoveContainer(c.Id, true, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Runner) newClient() (*dockerclient.DockerClient, error) {
	hh := host.Host(*h)
	url, err := hh.GetURL()
	if err != nil {
		return nil, err
	}

	tlsConfig, err := loadTLSConfig(h.HostOptions.AuthOptions.CaCertPath,
		h.HostOptions.AuthOptions.ClientCertPath,
		h.HostOptions.AuthOptions.ClientKeyPath,
		true)
	if err != nil {
		return nil, err
	}

	cli, err := dockerclient.NewDockerClient(url, tlsConfig)
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func (h *Runner) FindContainerByImage(storeType map[string]int) (string, error) {
	// return something like consul://ip:port
	cli, err := h.newClient()
	if err != nil {
		return "", err
	}

	containers, err := cli.ListContainers(false, false, "")
	for image, port := range storeType {
		url, err := h.findContainerByImage(containers, image, port)
		if err == nil {
			return url, nil
		}
	}

	return "", errors.New("No container found.")
}

func (h *Runner) findContainerByImage(containers []dockerclient.Container, image string, port int) (string, error) {
	for _, c := range containers {
		if strings.HasSuffix(c.Image, "/"+image) || strings.Contains(c.Image, "/"+image+":") {
			ip, err := h.Driver.GetIP()
			if err != nil {
				return "", err
			}

			protocol := image
			if image == "zookeeper" {
				protocol = "zk"
			}

			// name := h.Name
			// host port?
			if len(c.Ports) == 0 {
				return fmt.Sprintf("%s://%s:%d", protocol, ip, port), nil
			} else {
				for _, p := range c.Ports {
					if p.PublicPort == port {
						return fmt.Sprintf("%s://%s:%d", protocol, ip, port), nil
					}
				}
			}
		}
	}

	return "", errors.New("No container found.")
}
