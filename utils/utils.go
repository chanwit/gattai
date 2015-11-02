package utils

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

var (
	BaseDir = os.Getenv("MACHINE_STORAGE_PATH")
)

func GetBaseDir() string {
	if BaseDir == "" {
		BaseDir = filepath.Join(".gattai", "machine")
	}
	return BaseDir
}

func GetMachineDir() string {
	return filepath.Join(GetBaseDir(), "machines")
}

func GetMachineCertDir() string {
	return filepath.Join(GetBaseDir(), "certs")
}

func GetMachineCacheDir() string {
	return filepath.Join(GetBaseDir(), "cache")
}

func ReadFile(file string) ([]byte, error) {
	result := []byte{}

	log.Debugf("Opening file: %s", file)

	if file == "-" {
		if bytes, err := ioutil.ReadAll(os.Stdin); err != nil {
			log.Debugf("Failed to read file from stdin: %v", err)
			return nil, err
		} else {
			result = bytes
		}
	} else if file != "" {
		if bytes, err := ioutil.ReadFile(file); os.IsNotExist(err) {
			log.Debugf("Failed to find %s", file)
			return nil, err
		} else if err != nil {
			log.Debugf("Failed to open %s", file)
			return nil, err
		} else {
			result = bytes
		}
	}

	return result, nil
}

func Generate(pattern string) []string {
	re, _ := regexp.Compile(`\[(.+):(.+)\]`)
	submatch := re.FindStringSubmatch(pattern)
	if submatch == nil {
		return []string{pattern}
	}

	from, err := strconv.Atoi(submatch[1])
	if err != nil {
		return []string{pattern}
	}
	to, err := strconv.Atoi(submatch[2])
	if err != nil {
		return []string{pattern}
	}

	template := re.ReplaceAllString(pattern, "%d")

	var result []string
	for val := from; val <= to; val++ {
		entry := fmt.Sprintf(template, val)
		result = append(result, entry)
	}

	return result
}

func IncAddress(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
