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
			fmt.Println(err)
			os.Exit(1)
		}
		err := cmd.File.Execute()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(`{"file":"ok"}`)
		os.Exit(0)
	}

}
