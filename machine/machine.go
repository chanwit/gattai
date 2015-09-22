package machine

import (
	"os"
	"path/filepath"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/log"
	"github.com/docker/machine/utils"
)

func GetCertInfo() libmachine.CertPathInfo {
	caCertPath := filepath.Join(utils.GetMachineCertDir(), "ca.pem")
	caKeyPath := filepath.Join(utils.GetMachineCertDir(), "ca-key.pem")
	clientCertPath := filepath.Join(utils.GetMachineCertDir(), "cert.pem")
	clientKeyPath := filepath.Join(utils.GetMachineCertDir(), "key.pem")

	certInfo := libmachine.CertPathInfo{
		CaCertPath:     caCertPath,
		CaKeyPath:      caKeyPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
	}

	return certInfo
}

func GetProvider(storagePath string, certInfo libmachine.CertPathInfo) (*libmachine.Provider, error) {
	store := libmachine.NewFilestore(
		storagePath,
		certInfo.CaCertPath,
		certInfo.CaKeyPath)

	provider, err := libmachine.New(store)

	return provider, err
}

func SetupCertificates(caCertPath, caKeyPath, clientCertPath, clientKeyPath string) error {
	org := utils.GetUsername()
	bits := 2048

	if _, err := os.Stat(utils.GetMachineCertDir()); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(utils.GetMachineCertDir(), 0700); err != nil {
				log.Fatalf("Error creating machine config dir: %s", err)
			}
		} else {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		log.Infof("Creating CA: %s", caCertPath)

		// check if the key path exists; if so, error
		if _, err := os.Stat(caKeyPath); err == nil {
			log.Fatalf("The CA key already exists.  Please remove it or specify a different key/cert.")
		}

		if err := utils.GenerateCACertificate(caCertPath, caKeyPath, org, bits); err != nil {
			log.Infof("Error generating CA certificate: %s", err)
		}
	}

	if _, err := os.Stat(clientCertPath); os.IsNotExist(err) {
		log.Infof("Creating client certificate: %s", clientCertPath)

		if _, err := os.Stat(utils.GetMachineCertDir()); err != nil {
			if os.IsNotExist(err) {
				if err := os.Mkdir(utils.GetMachineCertDir(), 0700); err != nil {
					log.Fatalf("Error creating machine client cert dir: %s", err)
				}
			} else {
				log.Fatal(err)
			}
		}

		// check if the key path exists; if so, error
		if _, err := os.Stat(clientKeyPath); err == nil {
			log.Fatalf("The client key already exists.  Please remove it or specify a different key/cert.")
		}

		if err := utils.GenerateCert([]string{""}, clientCertPath, clientKeyPath, caCertPath, caKeyPath, org, bits); err != nil {
			log.Fatalf("Error generating client certificate: %s", err)
		}
	}

	return nil
}
