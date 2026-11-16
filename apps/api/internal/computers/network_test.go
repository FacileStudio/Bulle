package computers

import (
	"context"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/worker"
	"github.com/FacileStudio/tronc/errors"
)

type fakeNetworks struct {
	mu     sync.Mutex
	modes  map[int64]string
	allows map[int64][]string
}

func (f *fakeNetworks) ComputerNetwork(_ context.Context, userID int64) (string, []string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.modes[userID], f.allows[userID], nil
}

func (f *fakeNetworks) SaveComputerNetwork(_ context.Context, userID int64, network string, allow []string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.allows == nil {
		f.allows = map[int64][]string{}
	}
	f.modes[userID] = network
	f.allows[userID] = allow
	return nil
}

func (f *fakeWorker) SetNet(_ context.Context, id, net string, allow []string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, "net:"+net)
	f.boxes[id].Net = net
	f.boxes[id].NetAllow = allow
	return nil
}

func networkService(t *testing.T, modes map[int64]string) (*Service, *fakeWorker) {
	t.Helper()
	service, fw, _ := newTestService(t)
	service.Networks = &fakeNetworks{modes: modes}
	return service, fw
}

func TestOfflineUserComputerIsCreatedRestricted(t *testing.T) {
	service, fw := networkService(t, map[int64]string{3: NetworkOffline})
	computer, err := service.Ensure(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if got := fw.boxes[computer.BoxID].Net; got != NetworkOffline {
		t.Fatalf("created box net = %q, want offline", got)
	}
}

func TestOpenUserCreateOmitsTheNet(t *testing.T) {
	service, fw := networkService(t, map[int64]string{})
	computer, err := service.Ensure(context.Background(), 4)
	if err != nil {
		t.Fatal(err)
	}
	if got := fw.boxes[computer.BoxID].Net; got != "" {
		t.Fatalf("open create sent net %q, want it omitted", got)
	}
	status, err := service.Status(context.Background(), 4)
	if err != nil || status.Network != NetworkOpen {
		t.Fatalf("status = %+v err = %v", status, err)
	}
}

func TestSetNetworkRestartsARunningComputer(t *testing.T) {
	service, fw := networkService(t, map[int64]string{})
	if _, err := service.Ensure(context.Background(), 5); err != nil {
		t.Fatal(err)
	}
	fw.calls = nil
	computer, err := service.SetNetwork(context.Background(), 5, NetworkOffline, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(fw.calls, ","); got != "net:offline,stop,start" {
		t.Fatalf("calls = %s", got)
	}
	if computer.Network != NetworkOffline {
		t.Fatalf("computer = %+v", computer)
	}
	fw.calls = nil
	if _, err := service.SetNetwork(context.Background(), 5, NetworkOffline, nil); err != nil || len(fw.calls) != 0 {
		t.Fatalf("same mode touched the worker: %v err = %v", fw.calls, err)
	}
}

func TestStartAlignsAStoppedBoxWithTheChoice(t *testing.T) {
	modes := map[int64]string{}
	service, fw := networkService(t, modes)
	fw.boxes["old"] = &worker.Box{ID: "old", Owner: "6", Status: worker.StatusStopped}
	modes[6] = NetworkOffline
	if _, err := service.Launch(context.Background(), 6); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(fw.calls, ","); !strings.HasPrefix(got, "net:offline,start") {
		t.Fatalf("calls = %s", got)
	}
}

func TestSetNetworkRejectsUnknownModes(t *testing.T) {
	service, _ := networkService(t, map[int64]string{})
	if _, err := service.SetNetwork(context.Background(), 1, "lan", nil); err == nil {
		t.Fatal("want an error for an unknown mode")
	}
	computer, err := service.SetNetwork(context.Background(), 1, NetworkOffline, nil)
	if err != nil || computer.Status != StatusAbsent || computer.Network != NetworkOffline {
		t.Fatalf("absent computer = %+v err = %v", computer, err)
	}
}

func TestSetNetworkAllowlistSavesAndSendsTheHosts(t *testing.T) {
	service, fw := networkService(t, map[int64]string{})
	if _, err := service.Ensure(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
	fw.calls = nil
	computer, err := service.SetNetwork(context.Background(), 7, NetworkAllowlist, []string{" API.Example.com ", "*.cdn.example.org", "git.example.net:9418", "api.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"api.example.com", "*.cdn.example.org", "git.example.net:9418"}
	if computer.Network != NetworkAllowlist || !slices.Equal(computer.NetAllow, want) {
		t.Fatalf("computer = %+v", computer)
	}
	if got := strings.Join(fw.calls, ","); got != "net:allowlist,stop,start" {
		t.Fatalf("calls = %s", got)
	}
	if got := fw.boxes[computer.BoxID].NetAllow; !slices.Equal(got, want) {
		t.Fatalf("worker hosts = %v", got)
	}
	fw.calls = nil
	if _, err := service.SetNetwork(context.Background(), 7, NetworkAllowlist, []string{"git.example.net:9418", "*.cdn.example.org", "api.example.com"}); err != nil || len(fw.calls) != 0 {
		t.Fatalf("same hosts in another order touched the worker: %v err = %v", fw.calls, err)
	}
	computer, err = service.SetNetwork(context.Background(), 7, NetworkAllowlist, nil)
	if err != nil || len(computer.NetAllow) != 0 || len(fw.boxes[computer.BoxID].NetAllow) != 0 {
		t.Fatalf("omitting the hosts must clear them: %+v err = %v", computer, err)
	}
}

func TestSetNetworkRejectsBadHostsBeforeTheWorker(t *testing.T) {
	service, fw := networkService(t, map[int64]string{})
	if _, err := service.Ensure(context.Background(), 8); err != nil {
		t.Fatal(err)
	}
	fw.calls = nil
	tooMany := make([]string, MaxNetAllow+1)
	for i := range tooMany {
		tooMany[i] = "h" + strconv.Itoa(i) + ".example.com"
	}
	cases := map[string][]string{
		"ip literal":         {"10.0.0.1"},
		"ipv6":               {"[::1]:443"},
		"single label":       {"intranet"},
		"bare wildcard":      {"*.com"},
		"bad port":           {"example.com:0"},
		"port too large":     {"example.com:70000"},
		"underscore":         {"my_host.example.com"},
		"leading dash":       {"-a.example.com"},
		"numeric tld":        {"example.123"},
		"url":                {"https://example.com"},
		"too many":           tooMany,
		"one bad among good": {"example.com", "bad host.example.com"},
	}
	for name, hosts := range cases {
		_, err := service.SetNetwork(context.Background(), 8, NetworkAllowlist, hosts)
		if errors.Status(err) != http.StatusBadRequest {
			t.Errorf("%s: err = %v, want a 400", name, err)
		}
	}
	if _, err := service.SetNetwork(context.Background(), 8, NetworkOpen, []string{"example.com"}); err == nil {
		t.Error("hosts with the open mode must be refused")
	}
	if len(fw.calls) != 0 {
		t.Fatalf("a refused request reached the worker: %v", fw.calls)
	}
}

func TestAllowlistIsTheDefaultOnlyForNewComputersWhenTheInstanceAsks(t *testing.T) {
	modes := map[int64]string{}
	service, fw := networkService(t, modes)
	service.config.DefaultNetwork = NetworkAllowlist
	fw.boxes["old"] = &worker.Box{ID: "old", Owner: "9", Status: worker.StatusStopped}
	if _, err := service.Launch(context.Background(), 9); err != nil {
		t.Fatal(err)
	}
	if got := fw.boxes["old"].Net; got != "" {
		t.Fatalf("an existing computer was moved to %q", got)
	}
	status, err := service.Status(context.Background(), 9)
	if err != nil || status.Network != NetworkOpen {
		t.Fatalf("existing computer status = %+v err = %v", status, err)
	}
	computer, err := service.Ensure(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if got := fw.boxes[computer.BoxID].Net; got != NetworkAllowlist {
		t.Fatalf("new computer net = %q, want allowlist", got)
	}
	if modes[10] != NetworkAllowlist {
		t.Fatalf("the default was not kept as the choice: %q", modes[10])
	}
	service.config.DefaultNetwork = NetworkOpen
	status, err = service.Status(context.Background(), 10)
	if err != nil || status.Network != NetworkAllowlist {
		t.Fatalf("a later change of the default moved the computer: %+v err = %v", status, err)
	}
}

func TestTheDefaultStaysOpenWithoutTheFlag(t *testing.T) {
	service, fw := networkService(t, map[int64]string{})
	computer, err := service.Ensure(context.Background(), 11)
	if err != nil || fw.boxes[computer.BoxID].Net != "" {
		t.Fatalf("box = %+v err = %v", fw.boxes[computer.BoxID], err)
	}
}

func TestNetworkOfReadsTheChoiceWithoutTheWorker(t *testing.T) {
	service, fw := networkService(t, map[int64]string{12: NetworkAllowlist})
	service.Networks.(*fakeNetworks).allows = map[int64][]string{12: {"a.example.com"}}
	mode, hosts := service.NetworkOf(context.Background(), 12)
	if mode != NetworkAllowlist || !slices.Equal(hosts, []string{"a.example.com"}) || len(fw.calls) != 0 {
		t.Fatalf("mode = %q hosts = %v calls = %v", mode, hosts, fw.calls)
	}
	if mode, _ := service.NetworkOf(context.Background(), 13); mode != NetworkOpen {
		t.Fatalf("a person who never chose reads %q, want open", mode)
	}
}
