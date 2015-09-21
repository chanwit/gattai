package client

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/parsers/filters"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/log"
	"github.com/docker/machine/utils"
	"github.com/skarademir/naturalsort"
	"sort"
)

// FilterOptions -
type FilterOptions struct {
	SwarmName  []string
	DriverName []string
	State      []string
	Name       []string
}

func sortHostListItemsByName(items []libmachine.HostListItem) {
	m := make(map[string]libmachine.HostListItem, len(items))
	s := make([]string, len(items))
	for i, v := range items {
		name := strings.ToLower(v.Name)
		m[name] = v
		s[i] = name
	}
	sort.Sort(naturalsort.NaturalSort(s))
	for i, v := range s {
		items[i] = m[v]
	}
}

func (cli *DockerCli) CmdLs(args ...string) error {
	cmd := Cli.Subcmd("ls", []string{}, "List machines", false)

	psFilterArgs := filters.Args{}
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "List quietly")
	flFilter := opts.NewListOpts(nil)

	cmd.Var(&flFilter, []string{"f", "-filter"}, "Filter output based on conditions provided")

	cmd.ParseFlags(args, true)

	var err error
	// Consolidate all filter flags, and sanity check them.
	// They'll get processed in the daemon/server.
	for _, f := range flFilter.GetAll() {
		if psFilterArgs, err = filters.ParseFlag(f, psFilterArgs); err != nil {
			return err
		}
	}

	options := FilterOptions{}
	if len(psFilterArgs) > 0 {
		options.SwarmName = psFilterArgs["swarm"]
		options.DriverName = psFilterArgs["driver"]
		options.State = psFilterArgs["state"]
		options.Name = psFilterArgs["name"]
	}

	certInfo := GetCertInfo()
	provider, err := GetProvider(utils.GetBaseDir(), certInfo)
	if err != nil {
		log.Fatal(err)
	}

	hostList, err := provider.List()
	if err != nil {
		log.Fatal(err)
	}

	hostList = filterHosts(hostList, options)

	// Just print out the names if we're being quiet
	if *quiet {
		for _, host := range hostList {
			fmt.Println(host.Name)
		}
		return nil
	}

	swarmMasters := make(map[string]string)
	swarmInfo := make(map[string]string)

	w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tACTIVE\tDRIVER\tSTATE\tURL\tSWARM")

	for _, host := range hostList {
		swarmOptions := host.HostOptions.SwarmOptions
		if swarmOptions.Master {
			swarmMasters[swarmOptions.Discovery] = host.Name
		}

		if swarmOptions.Discovery != "" {
			swarmInfo[host.Name] = swarmOptions.Discovery
		}
	}

	items := libmachine.GetHostListItems(hostList)

	sortHostListItemsByName(items)

	for _, item := range items {
		activeString := "-"
		if item.Active {
			activeString = "*"
		}

		swarmInfo := ""

		if item.SwarmOptions.Discovery != "" {
			swarmInfo = swarmMasters[item.SwarmOptions.Discovery]
			if item.SwarmOptions.Master {
				swarmInfo = fmt.Sprintf("%s (master)", swarmInfo)
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Name, activeString, item.DriverName, item.State, item.URL, swarmInfo)
	}

	w.Flush()
	return nil
}

func parseFilters(filters []string) (FilterOptions, error) {
	options := FilterOptions{}
	for _, f := range filters {
		kv := strings.SplitN(f, "=", 2)
		if len(kv) != 2 {
			return options, errors.New("Unsupported filter syntax.")
		}
		key, value := kv[0], kv[1]

		switch key {
		case "swarm":
			options.SwarmName = append(options.SwarmName, value)
		case "driver":
			options.DriverName = append(options.DriverName, value)
		case "state":
			options.State = append(options.State, value)
		case "name":
			options.Name = append(options.Name, value)
		default:
			return options, fmt.Errorf("Unsupported filter key '%s'", key)
		}
	}
	return options, nil
}

func filterHosts(hosts []*libmachine.Host, filters FilterOptions) []*libmachine.Host {
	if len(filters.SwarmName) == 0 &&
		len(filters.DriverName) == 0 &&
		len(filters.State) == 0 &&
		len(filters.Name) == 0 {
		return hosts
	}

	filteredHosts := []*libmachine.Host{}
	swarmMasters := getSwarmMasters(hosts)

	for _, h := range hosts {
		if filterHost(h, filters, swarmMasters) {
			filteredHosts = append(filteredHosts, h)
		}
	}
	return filteredHosts
}

func getSwarmMasters(hosts []*libmachine.Host) map[string]string {
	swarmMasters := make(map[string]string)
	for _, h := range hosts {
		swarmOptions := h.HostOptions.SwarmOptions
		if swarmOptions != nil && swarmOptions.Master {
			swarmMasters[swarmOptions.Discovery] = h.Name
		}
	}
	return swarmMasters
}

func filterHost(host *libmachine.Host, filters FilterOptions, swarmMasters map[string]string) bool {
	swarmMatches := matchesSwarmName(host, filters.SwarmName, swarmMasters)
	driverMatches := matchesDriverName(host, filters.DriverName)
	stateMatches := matchesState(host, filters.State)
	nameMatches := matchesName(host, filters.Name)

	return swarmMatches && driverMatches && stateMatches && nameMatches
}

func matchesSwarmName(host *libmachine.Host, swarmNames []string, swarmMasters map[string]string) bool {
	if len(swarmNames) == 0 {
		return true
	}
	for _, n := range swarmNames {
		if host.HostOptions.SwarmOptions != nil {
			if n == swarmMasters[host.HostOptions.SwarmOptions.Discovery] {
				return true
			}
		}
	}
	return false
}

func matchesDriverName(host *libmachine.Host, driverNames []string) bool {
	if len(driverNames) == 0 {
		return true
	}
	for _, n := range driverNames {
		if host.DriverName == n {
			return true
		}
	}
	return false
}

func matchesState(host *libmachine.Host, states []string) bool {
	if len(states) == 0 {
		return true
	}
	for _, n := range states {
		s, err := host.Driver.GetState()
		if err != nil {
			log.Warn(err)
		}
		if n == s.String() {
			return true
		}
	}
	return false
}

func matchesName(host *libmachine.Host, names []string) bool {
	if len(names) == 0 {
		return true
	}
	for _, n := range names {
		r, err := regexp.Compile(n)
		if err != nil {
			log.Fatal(err)
		}
		if r.MatchString(host.Driver.GetMachineName()) {
			return true
		}
	}
	return false
}
