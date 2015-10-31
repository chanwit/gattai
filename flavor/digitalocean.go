package flavor

import "github.com/chanwit/gattai/machine"

// Digital Ocean Flavors

var (
	DigitalOcean_2G = machine.Machine{
		Driver: "digitalocean",
		Options: map[string]interface{}{
			"digitalocean-image":        "ubuntu-14-04-x64",
			"digitalocean-region":       "nyc3",
			"digitalocean-size":         "2gb",
			"digitalocean-access-token": "$DIGITALOCEAN_ACCESS_TOKEN",
			"engine-install-url":        "https://get.docker.com",
		},
	}

	DigitalOcean_2G_Exp = machine.Machine{
		Driver: "digitalocean",
		Options: map[string]interface{}{
			"digitalocean-image":        "debian-8-x64",
			"digitalocean-region":       "nyc3",
			"digitalocean-size":         "2gb",
			"digitalocean-access-token": "$DIGITALOCEAN_ACCESS_TOKEN",
			"engine-install-url":        "https://experimental.docker.com",
		},
	}

	DigitalOcean_2G_Cluster = map[string]machine.Machine{
		"node": machine.Machine{
			Driver: "digitalocean",
			Options: map[string]interface{}{
				"digitalocean-image":        "debian-8-x64",
				"digitalocean-region":       "nyc3",
				"digitalocean-size":         "2gb",
				"digitalocean-access-token": "$DIGITALOCEAN_ACCESS_TOKEN",
				"engine-install-url":        "https://experimental.docker.com",
			},
			PostProvision: []string{
				"docker network create -d overlay multihost",
			},
		},
		"master": machine.Machine{
			Driver:    "digitalocean",
			Instances: 1,
			Options: map[string]interface{}{
				"digitalocean-image":        "debian-8-x64",
				"digitalocean-region":       "nyc3",
				"digitalocean-size":         "2gb",
				"digitalocean-access-token": "$DIGITALOCEAN_ACCESS_TOKEN",
				"engine-install-url":        "https://experimental.docker.com",
			},
			PostProvision: []string{
				"docker run -d -p 8400:8400 -p 8500:8500 -p 8600:53/udp progrium/consul --server -bootstrap-expect 1",
			},
		},
	}
)
