package machine

import (
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine/cert"
	// utils "github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/persist"
)

func GetCertInfo() cert.CertPathInfo {
	caCertPath := filepath.Join(mcndirs.GetMachineCertDir(), "ca.pem")
	caKeyPath := filepath.Join(mcndirs.GetMachineCertDir(), "ca-key.pem")
	clientCertPath := filepath.Join(mcndirs.GetMachineCertDir(), "cert.pem")
	clientKeyPath := filepath.Join(mcndirs.GetMachineCertDir(), "key.pem")

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
