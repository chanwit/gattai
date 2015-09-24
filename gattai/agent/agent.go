package main

import (
	"fmt"
	"os"
	"strings"

	"encoding/json"
)

type Cmd struct {
	File FileCmd `json:"file"`
}

func main() {
	code := os.Args[1]
	if strings.HasPrefix(code, `{"file":`) {
		cmd := Cmd{}
		if err := json.Unmarshal([]byte(code), &cmd); err != nil {
			panic(err)
		}
		err := cmd.File.Execute()
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("OK")
}
