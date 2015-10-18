package machine

import (
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/chanwit/gattai/utils"
	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/persist"
)

func GetCertInfo() cert.CertPathInfo {
	caCertPath := filepath.Join(utils.GetMachineCertDir(), "ca.pem")
	caKeyPath := filepath.Join(utils.GetMachineCertDir(), "ca-key.pem")
	clientCertPath := filepath.Join(utils.GetMachineCertDir(), "cert.pem")
	clientKeyPath := filepath.Join(utils.GetMachineCertDir(), "key.pem")

	certInfo := cert.CertPathInfo{
		CaCertPath:       caCertPath,
		CaPrivateKeyPath: caKeyPath,
		ClientCertPath:   clientCertPath,
		ClientKeyPath:    clientKeyPath,
	}

	return certInfo
}

func GetDefaultStore(storagePath string) *persist.Filestore {
	// storagePath = ".gattai/machine"
	log.Debug("GetDefaultStore")
	certsDir := filepath.Join(storagePath, "certs")
	return &persist.Filestore{
		Path:             storagePath,
		CaCertPath:       certsDir,
		CaPrivateKeyPath: certsDir,
	}
}
