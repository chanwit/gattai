package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"

	"github.com/docker/machine/log"
)

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
