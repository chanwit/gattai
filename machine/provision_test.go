package machine

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type state int

const (
	KEY = state(iota)
	VALUE
	IN_PARENT
	IN_PARENT2
	SPACE
)

func parseAttributes(s string) map[string]string {
	st := KEY
	i := 0
	str := ""
	key := ""

	result := make(map[string]string)
	for {
		ch := s[i]
		switch ch {
		case '=':
			if st == KEY {
				key = str
				str = ""
				st = VALUE
			}
		case '\'':
			if st == VALUE {
				st = IN_PARENT2
			} else if st == IN_PARENT2 {
				st = VALUE
			} else {
				str = str + string(ch)
			}
		case '"':
			if st == VALUE {
				st = IN_PARENT
			} else if st == IN_PARENT {
				st = VALUE
			} else {
				str = str + string(ch)
			}
		case ' ':
			if st == VALUE {
				result[key] = str
				st = SPACE
			} else {
				str = str + string(ch)
			}
		default:
			if st == SPACE {
				st = KEY
				str = ""
			}
			str = str + string(ch)
		}
		i++
		if i >= len(s) {
			if st != SPACE {
				result[key] = str
			}
			break
		}
	}

	return result
}

func TestParsing(t *testing.T) {
	p, err := parseProvision([]byte(`---
machines:
  fake:
    driver: none
    instances: 1
    options:
      url: "tcp://1.2.3.4:2375"
    commands:
      - file: src=test.txt dest=/ state=present
      - file: src=test.txt dest="/test a" state=present
    `))
	assert.NoError(t, err)
	file0 := p.Machines["fake"].Commands[0]["file"]
	assert.Equal(t, file0, "src=test.txt dest=/ state=present")

	attr := parseAttributes(file0)
	assert.Equal(t, attr["src"], "test.txt")
	assert.Equal(t, attr["dest"], "/")
	assert.Equal(t, attr["state"], "present")

	file1 := p.Machines["fake"].Commands[1]["file"]
	assert.Equal(t, file1, "src=test.txt dest=\"/test a\" state=present")
	attr = parseAttributes(file1)
	assert.Equal(t, attr["src"], "test.txt")
	assert.Equal(t, attr["dest"], "/test a")
	assert.Equal(t, attr["state"], "present")
}
