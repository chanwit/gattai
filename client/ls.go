package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/chanwit/gattai/machine"
	"github.com/chanwit/gattai/machine/driverfactory"
	"github.com/chanwit/gattai/utils"

	Cli "github.com/docker/docker/cli"

	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/parsers/filters"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/persist"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"

	"github.com/skarademir/naturalsort"
)

var (
	stateTimeoutDuration = 10 * time.Second
)

// FilterOptions -
type FilterOptions struct {
	SwarmName  []string
	DriverName []string
	State      []string
	Name       []string
}

type HostListItem struct {
	Name         string
	Active       bool
	DriverName   string
	State        state.State
	URL          string
	SwarmOptions *swarm.SwarmOptions
}

// func cmdLs(c *cli.Context) {
func DoLs(cli interface{}, args ...string) error {
	// func (cli *DockerCli) CmdLs(args ...string) error {
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

	store := machine.GetDefaultStore(utils.GetBaseDir())
	hostList, err := listHosts(store, utils.GetBaseDir()) // TODO
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
	fmt.Fprintln(w, "NAME\tDRIVER\tSTATE\tURL")

	for _, host := range hostList {
		swarmOptions := host.HostOptions.SwarmOptions
		if swarmOptions.Master {
			swarmMasters[swarmOptions.Discovery] = host.Name
		}

		if swarmOptions.Discovery != "" {
			swarmInfo[host.Name] = swarmOptions.Discovery
		}
	}

	items := getHostListItems(hostList)

	sortHostListItemsByName(items)

	for _, item := range items {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", item.Name, item.DriverName, item.State, item.URL)
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

func filterHosts(hosts []*host.Host, filters FilterOptions) []*host.Host {
	if len(filters.SwarmName) == 0 &&
		len(filters.DriverName) == 0 &&
		len(filters.State) == 0 &&
		len(filters.Name) == 0 {
		return hosts
	}

	filteredHosts := []*host.Host{}
	swarmMasters := getSwarmMasters(hosts)

	for _, h := range hosts {
		if filterHost(h, filters, swarmMasters) {
			filteredHosts = append(filteredHosts, h)
		}
	}
	return filteredHosts
}

func getSwarmMasters(hosts []*host.Host) map[string]string {
	swarmMasters := make(map[string]string)
	for _, h := range hosts {
		swarmOptions := h.HostOptions.SwarmOptions
		if swarmOptions != nil && swarmOptions.Master {
			swarmMasters[swarmOptions.Discovery] = h.Name
		}
	}
	return swarmMasters
}

func filterHost(host *host.Host, filters FilterOptions, swarmMasters map[string]string) bool {
	swarmMatches := matchesSwarmName(host, filters.SwarmName, swarmMasters)
	driverMatches := matchesDriverName(host, filters.DriverName)
	stateMatches := matchesState(host, filters.State)
	nameMatches := matchesName(host, filters.Name)

	return swarmMatches && driverMatches && stateMatches && nameMatches
}

func matchesSwarmName(host *host.Host, swarmNames []string, swarmMasters map[string]string) bool {
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

func matchesDriverName(host *host.Host, driverNames []string) bool {
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

func matchesState(host *host.Host, states []string) bool {
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

func matchesName(host *host.Host, names []string) bool {
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

func getActiveHost(store persist.Store) (*host.Host, error) {
	hosts, err := store.List()
	if err != nil {
		return nil, err
	}

	hostListItems := getHostListItems(hosts)

	for _, item := range hostListItems {
		if item.Active {
			h, err := store.Load(item.Name)
			if err != nil {
				return nil, err
			}
			return h, nil
		}
	}

	return nil, errors.New("Active host not found")
}

func attemptGetHostState(h *host.Host, stateQueryChan chan<- HostListItem) {
	currentState, err := h.Driver.GetState()
	if err != nil {
		log.Errorf("error getting state for host %s: %s", h.Name, err)
	}

	url, err := h.GetURL()
	if err != nil {
		if err == drivers.ErrHostIsNotRunning {
			url = ""
		} else {
			log.Errorf("error getting URL for host %s: %s", h.Name, err)
		}
	}

	active, err := isActive(h)
	if err != nil {
		log.Errorf("error determining if host is active for host %s: %s",
			h.Name, err)
	}

	stateQueryChan <- HostListItem{
		Name:         h.Name,
		Active:       active,
		DriverName:   h.Driver.DriverName(),
		State:        currentState,
		URL:          url,
		SwarmOptions: h.HostOptions.SwarmOptions,
	}
}

func getHostState(h *host.Host, hostListItemsChan chan<- HostListItem) {
	// This channel is used to communicate the properties we are querying
	// about the host in the case of a successful read.
	stateQueryChan := make(chan HostListItem)

	go attemptGetHostState(h, stateQueryChan)

	select {
	// If we get back useful information, great.  Forward it straight to
	// the original parent channel.
	case hli := <-stateQueryChan:
		hostListItemsChan <- hli

	// Otherwise, give up after a predetermined duration.
	case <-time.After(stateTimeoutDuration):
		hostListItemsChan <- HostListItem{
			Name:       h.Name,
			DriverName: h.Driver.DriverName(),
			State:      state.Timeout,
		}
	}
}

func getHostListItems(hostList []*host.Host) []HostListItem {
	hostListItems := []HostListItem{}
	hostListItemsChan := make(chan HostListItem)

	for _, h := range hostList {
		go getHostState(h, hostListItemsChan)
	}

	for range hostList {
		hostListItems = append(hostListItems, <-hostListItemsChan)
	}

	close(hostListItemsChan)
	return hostListItems
}

// IsActive provides a single function for determining if a host is active
// based on both the url and if the host is stopped.
func isActive(h *host.Host) (bool, error) {
	currentState, err := h.Driver.GetState()

	if err != nil {
		log.Errorf("error getting state for host %s: %s", h.Name, err)
		return false, err
	}

	url, err := h.GetURL()

	if err != nil {
		if err == drivers.ErrHostIsNotRunning {
			url = ""
		} else {
			log.Errorf("error getting URL for host %s: %s", h.Name, err)
			return false, err
		}
	}

	dockerHost := os.Getenv("DOCKER_HOST")

	notStopped := currentState != state.Stopped
	correctURL := url == dockerHost

	isActive := notStopped && correctURL

	return isActive, nil
}

func sortHostListItemsByName(items []HostListItem) {
	m := make(map[string]HostListItem, len(items))
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

func loadHost(store persist.Store, hostName string, storePath string) (*host.Host, error) {
	h, err := store.Load(hostName)
	if err != nil {
		return nil, fmt.Errorf("Loading host from store failed: %s", err)
	}

	d, err := driverfactory.NewDriver(h.DriverName, h.Name, storePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(h.RawDriver, &d)
	if err != nil {
		return nil, err
	}
	h.Driver = d

	return h, nil
}

func listHosts(store persist.Store, storePath string) ([]*host.Host, error) {
	cliHosts := []*host.Host{}

	hosts, err := store.List()
	if err != nil {
		return nil, fmt.Errorf("Error attempting to list hosts from store: %s", err)
	}

	for _, h := range hosts {
		h, err = loadHost(store, h.Name, storePath)
		if err != nil {
			return nil, err
		}
		cliHosts = append(cliHosts, h)
	}

	return cliHosts, nil
}
