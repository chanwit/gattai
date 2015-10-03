package machine

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/docker/machine/libmachine/mcnutils"
)

var defaultValues = map[string]interface{}{
	"amazonec2-region":         "us-east-1",
	"amazonec2-zone":           "a",
	"amazonec2-security-group": "docker-machine",
	"amazonec2-instance-type":  "t2.micro",
	"amazonec2-root-size":      16,
	"amazonec2-ssh-user":       "ubuntu",
	"amazonec2-spot-price":     "0.50",
	"amazonec2-ami":            "ami-615cb725",

	"azure-docker-port":              2376,
	"azure-docker-swarm-master-port": 3376,
	"azure-location":                 "West US",
	"azure-size":                     "Small",
	"azure-ssh-port":                 22,
	"azure-username":                 "ubuntu",

	"digitalocean-image":  "ubuntu-14-04-x64",
	"digitalocean-region": "nyc3",
	"digitalocean-size":   "512mb",

	"exoscale-instance-profile":  "small",
	"exoscale-disk-size":         50,
	"exoscale-image":             "ubuntu-14.04",
	"exoscale-availability-zone": "ch-gva-2",

	"generic-ssh-user": "root",
	"generic-ssh-key":  filepath.Join(mcnutils.GetHomeDir(), ".ssh", "id_rsa"),
	"generic-ssh-port": 22,

	"google-zone":         "us-central1-a",
	"google-machine-type": "f1-micro",
	"google-username":     "docker-user",
	"google-scopes":       "https://www.googleapis.com/auth/devstorage.read_only,https://www.googleapis.com/auth/logging.write",
	"google-disk-size":    "pd-standard",
	"google-disk-type":    10,

	"hyper-v-disk-size": 20000,
	"hyper-v-memory":    1024,

	"openstack-ssh-user":       "root",
	"openstack-ssh-port":       22,
	"openstack-active-timeout": 200,

	"rackspace-endpoint-type":  "publicURL",
	"rackspace-flavor-id":      "general1-1",
	"rackspace-ssh-user":       "root",
	"rackspace-ssh-port":       22,
	"rackspace-docker-install": "true",

	"softlayer-memory":       1024,
	"softlayer-disk-size":    0,
	"softlayer-region":       "dal01",
	"softlayer-cpu":          1,
	"softlayer-api-endpoint": "https://api.softlayer.com/rest/v3",
	"softlayer-image":        "UBUNTU_LATEST",

	"virtualbox-memory":                1024,
	"virtualbox-cpu-count":             1,
	"virtualbox-disk-size":             20000,
	"virtualbox-boot2docker-url":       "",
	"virtualbox-import-boot2docker-vm": "",
	"virtualbox-hostonly-cidr":         "192.168.99.1/24",
	"virtualbox-hostonly-nictype":      "82540EM",
	"virtualbox-hostonly-nicpromisc":   "deny",

	"vmwarefusion-cpu-count":    1,
	"vmwarefusion-memory-size":  1024,
	"vmwarefusion-disk-size":    20000,
	"vmwarefusion-ssh-user":     "docker",
	"vmwarefusion-ssh-password": "tcuser",

	"vmwarevcloudair-catalog":     "Public Catalog",
	"vmwarevcloudair-catalogitem": "Ubuntu Server 12.04 LTS (amd64 20150127)",
	"vmwarevcloudair-cpu-count":   1,
	"vmwarevcloudair-memory-size": 2048,
	"vmwarevcloudair-ssh-port":    22,
	"vmwarevcloudair-docker-port": 2376,

	"vmwarevsphere-cpu-count":   2,
	"vmwarevsphere-memory-size": 2048,
	"vmwarevsphere-disk-size":   20000,
}

type Options map[string]interface{}

func (o Options) String(key string) string {
	if value, exist := o[key]; exist {
		if s, ok := value.(string); ok {
			return os.ExpandEnv(s)
		}
	}

	if s, ok := defaultValues[key].(string); ok {
		return s
	}

	return ""
}

func (o Options) StringSlice(key string) []string {
	if value, exist := o[key]; exist {
		if s, ok := value.([]string); ok {
			result := []string{}
			for _, each := range s {
				result = append(result, os.ExpandEnv(each))
			}
			return result
		} else if s, ok := value.(string); ok {
			return []string{os.ExpandEnv(s)}
		}
	}

	if s, ok := defaultValues[key].(string); ok {
		return []string{s}
	}

	return []string{}
}

func (o Options) Int(key string) int {
	if value, exist := o[key]; exist {
		if i, ok := value.(int); ok {
			return i
		} else if s, ok := value.(string); ok {
			s = os.ExpandEnv(s)
			i, _ := strconv.Atoi(s)
			return i
		}
	}

	if i, ok := defaultValues[key].(int); ok {
		return i
	}

	return 0
}

func (o Options) Bool(key string) bool {
	if value, exist := o[key]; exist {

		if b, ok := value.(bool); ok {
			return b
		} else if s, ok := value.(string); ok {
			s = os.ExpandEnv(s)
			return s == "true" || s == "yes"
		}

	}

	if b, ok := defaultValues[key].(bool); ok {
		return b
	}

	return false
}
