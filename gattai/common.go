package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/cli"
	"github.com/docker/docker/cliconfig"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/tlsconfig"
)

const (
	defaultTrustKeyFile = "key.json"
	defaultCaFile       = "ca.pem"
	defaultKeyFile      = "key.pem"
	defaultCertFile     = "cert.pem"
)

var (
	daemonFlags *flag.FlagSet
	commonFlags = &cli.CommonFlags{FlagSet: new(flag.FlagSet)}

	dockerCertPath  string
	dockerTLSVerify bool
)

func checkPemFiles(envs map[string]string) bool {
	certPath := envs["DOCKER_CERT_PATH"]
	if _, err := os.Stat(filepath.Join(certPath, defaultCaFile)); err != nil {
		// No pem file
		// Reset cert path and tlsverify
		os.Setenv("DOCKER_CERT_PATH", "")
		os.Setenv("DOCKER_TLS_VERIFY", "")

		return false
	}

	return true
}

func initEnvs() {
	envs := make(map[string]string)
	bytes, err := readFile(".gattai/.active_host")
	err = yaml.Unmarshal(bytes, &envs)
	if err == nil {

		if checkPemFiles(envs) == false {
			return
		}
		// check pem file
		// reset

		if os.Getenv("DOCKER_HOST") == "" {
			os.Setenv("DOCKER_HOST", envs["DOCKER_HOST"])
		}
		if os.Getenv("DOCKER_CERT_PATH") == "" {
			os.Setenv("DOCKER_CERT_PATH", envs["DOCKER_CERT_PATH"])
		}
		if os.Getenv("DOCKER_TLS_VERIFY") == "" {
			os.Setenv("DOCKER_TLS_VERIFY", envs["DOCKER_TLS_VERIFY"])
		}
	}
}

func init() {
	initEnvs()
	dockerCertPath = os.Getenv("DOCKER_CERT_PATH")
	dockerTLSVerify = os.Getenv("DOCKER_TLS_VERIFY") != ""

	if dockerCertPath == "" {
		dockerCertPath = cliconfig.ConfigDir()
	}

	commonFlags.PostParse = postParseCommon

	cmd := commonFlags.FlagSet

	cmd.BoolVar(&commonFlags.Debug, []string{"D", "-debug"}, false, "Enable debug mode")
	cmd.StringVar(&commonFlags.LogLevel, []string{"l", "-log-level"}, "info", "Set the logging level")
	cmd.BoolVar(&commonFlags.TLS, []string{"-tls"}, false, "Use TLS; implied by --tlsverify")
	cmd.BoolVar(&commonFlags.TLSVerify, []string{"-tlsverify"}, dockerTLSVerify, "Use TLS and verify the remote")

	// TODO use flag flag.String([]string{"i", "-identity"}, "", "Path to libtrust key file")

	var tlsOptions tlsconfig.Options
	commonFlags.TLSOptions = &tlsOptions
	cmd.StringVar(&tlsOptions.CAFile, []string{"-tlscacert"}, filepath.Join(dockerCertPath, defaultCaFile), "Trust certs signed only by this CA")
	cmd.StringVar(&tlsOptions.CertFile, []string{"-tlscert"}, filepath.Join(dockerCertPath, defaultCertFile), "Path to TLS certificate file")
	cmd.StringVar(&tlsOptions.KeyFile, []string{"-tlskey"}, filepath.Join(dockerCertPath, defaultKeyFile), "Path to TLS key file")

	cmd.Var(opts.NewListOptsRef(&commonFlags.Hosts, opts.ValidateHost), []string{"H", "-host"}, "Daemon socket(s) to connect to")
}

func postParseCommon() {
	cmd := commonFlags.FlagSet

	if commonFlags.LogLevel != "" {
		lvl, err := logrus.ParseLevel(commonFlags.LogLevel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to parse logging level: %s\n", commonFlags.LogLevel)
			os.Exit(1)
		}
		logrus.SetLevel(lvl)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if commonFlags.Debug {
		os.Setenv("DEBUG", "1")
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Regardless of whether the user sets it to true or false, if they
	// specify --tlsverify at all then we need to turn on tls
	// TLSVerify can be true even if not set due to DOCKER_TLS_VERIFY env var, so we need to check that here as well
	if cmd.IsSet("-tlsverify") || commonFlags.TLSVerify {
		commonFlags.TLS = true
	}

	if !commonFlags.TLS {
		commonFlags.TLSOptions = nil
	} else {
		tlsOptions := commonFlags.TLSOptions
		tlsOptions.InsecureSkipVerify = !commonFlags.TLSVerify

		// Reset CertFile and KeyFile to empty string if the user did not specify
		// the respective flags and the respective default files were not found.
		if !cmd.IsSet("-tlscert") {
			if _, err := os.Stat(tlsOptions.CertFile); os.IsNotExist(err) {
				tlsOptions.CertFile = ""
			}
		}
		if !cmd.IsSet("-tlskey") {
			if _, err := os.Stat(tlsOptions.KeyFile); os.IsNotExist(err) {
				tlsOptions.KeyFile = ""
			}
		}
	}
}
