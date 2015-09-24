package main

import (
	"fmt"
	"os"
)

type FileCmd struct {
	Src *string `json:"src"`

	// path, alias: ['dest', 'name']
	Path *string `json:"path"`
	Dest *string `json:"dest"`
	Name *string `json:"name"`

	// file, link, directory, hard, touch, absent
	State *string `json:"state"`

	// apply only when state=directory
	Recurse bool `json:"recurse"`

	Force bool `json:"force"`
}

type FileCmdState int

const (
	FileState = FileCmdState(iota)
	LinkState
	DirectoryState
	HardState
	TouchState
	AbsentState
)

func (cmd FileCmd) GetSrc() string {
	if cmd.Src == nil {
		return ""
	}

	return *cmd.Src
}

func (cmd FileCmd) GetPath() string {
	if cmd.Path != nil {
		return *cmd.Path
	} else if cmd.Dest != nil {
		return *cmd.Dest
	} else if cmd.Name != nil {
		return *cmd.Name
	}

	panic("Path is required")
	return ""
}

func (cmd FileCmd) GetState() FileCmdState {

	states := map[string]FileCmdState{
		"file":      FileState,
		"link":      LinkState,
		"directory": DirectoryState,
		"hard":      HardState,
		"touch":     TouchState,
		"absent":    AbsentState,
	}

	stateStr := "file"
	if cmd.State != nil {
		stateStr = *cmd.State
	}

	value, exist := states[stateStr]
	if exist == false {
		panic("State specified is incorrect")
	}

	return value
}

func (cmd FileCmd) PreCondition() (bool, error) {
	switch cmd.GetState() {
	// case "file":
	// case "link":
	// case "directory":
	// case "hard":
	// case "touch":
	case AbsentState:
		path := cmd.GetPath()
		_, err := os.Stat(path)
		// err != nil means canot get state,
		// so there is no file
		if err != nil {
			return false, fmt.Errorf(`{"file":{"info": "%s is already absent"}}`, path)
		} else {
			// This means pre-condition is OK
			// Process to delete
			return true, nil
		}
	}

	return false, fmt.Errorf(`{"file":{"error": "pre-condition not met"}}`)
}

func (cmd FileCmd) Execute() error {
	if _, err := cmd.PreCondition(); err != nil {
		return err
	}

	switch cmd.GetState() {
	// case "file":
	// case "link":
	// case "directory":
	// case "hard":
	// case "touch":
	case AbsentState:
		err := os.Remove(cmd.GetPath())
		if err != nil {
			return err
		}
	}

	if _, err := cmd.PostCondition(); err != nil {
		return err
	}

	return nil
}

func (cmd FileCmd) PostCondition() (bool, error) {
	switch cmd.GetState() {
	case AbsentState:
		path := cmd.GetPath()
		_, err := os.Stat(path)
		if err != nil {
			// file is deleted
			// so return true
			return true, nil
		} else {
			return false, fmt.Errorf(`{"file":{"error": "cannot change %s to be absent"}}`, path)
		}
	}

	return false, fmt.Errorf(`{"file":{"error": "post-condition not met"}}`)
}
