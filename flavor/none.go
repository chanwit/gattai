package flavor

import "github.com/chanwit/gattai/machine"

// Digital Ocean Flavors

var (
	None = machine.Machine{
		Driver: "none",
		Options: map[string]interface{}{
			"url": "tcp://1.2.3.4:2376",
		},
	}
)
