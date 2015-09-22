package machine

import (
	"os"
	"strconv"
)

type Options map[string]interface{}

func (o Options) String(key string) string {
	value := o[key]
	if s, ok := value.(string); ok {
		return os.ExpandEnv(s)
	}
	return ""
}

func (o Options) StringSlice(key string) []string {
	value := o[key]
	if s, ok := value.([]string); ok {
		result := []string{}
		for _, each := range s {
			result = append(result, os.ExpandEnv(each))
		}
		return result
	} else if s, ok := value.(string); ok {
		return []string{os.ExpandEnv(s)}
	}

	return []string{}
}

func (o Options) Int(key string) int {
	value := o[key]
	if i, ok := value.(int); ok {
		return i
	} else if s, ok := value.(string); ok {
		s = os.ExpandEnv(s)
		i, _ := strconv.Atoi(s)
		return i
	}
	return 0
}

func (o Options) Bool(key string) bool {
	value := o[key]
	if b, ok := value.(bool); ok {
		return b
	} else if s, ok := value.(string); ok {
		s = os.ExpandEnv(s)
		return s == "true" || s == "yes"
	}
	return false
}
