package computers

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/worker"
)

type fakeNetworks struct {
	mu    sync.Mutex
	modes map[int64]string
}

func (f *fakeNetworks) ComputerNetwork(_ context.Context, userID int64) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.modes[userID], nil
}

func (f *fakeNetworks) SaveComputerNetwork(_ context.Context, userID int64, network string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.modes[userID] = network
	return nil
}

func (f *fakeWorker) SetNet(_ context.Context, id, net string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, "net:"+net)
	f.boxes[id].Net = net
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
	computer, err := service.SetNetwork(context.Background(), 5, NetworkOffline)
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
	if _, err := service.SetNetwork(context.Background(), 5, NetworkOffline); err != nil || len(fw.calls) != 0 {
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
	if _, err := service.SetNetwork(context.Background(), 1, "lan"); err == nil {
		t.Fatal("want an error for an unknown mode")
	}
	computer, err := service.SetNetwork(context.Background(), 1, NetworkOffline)
	if err != nil || computer.Status != StatusAbsent || computer.Network != NetworkOffline {
		t.Fatalf("absent computer = %+v err = %v", computer, err)
	}
}
