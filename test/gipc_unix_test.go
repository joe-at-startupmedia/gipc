//go:build !windows && !network

package gipc

import (
	"fmt"
	"github.com/joe-at-startupmedia/gipc"
	"os"
	"testing"
)

func TestUnixUnmask(t *testing.T) {
	scon := NewServerConfig("test_unmask")
	scon.UnmaskPermissions = true
	sc, err := gipc.StartServer(scon)
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	gipc.Sleep()

	info, err := os.Stat(sc.GetListener().Addr().String())
	if err != nil {
		t.Error(err)
	}
	got := fmt.Sprintf("%04o", info.Mode().Perm())
	want := "0777"

	if got != want {
		t.Errorf("Got %q, Wanted %q", got, want)
	}
}

// fails in network mode to due to reconnecting causing a hang
func TestUnixServerClose(t *testing.T) {

	sc, err := gipc.StartServer(NewServerConfig("test1010"))
	if err != nil {
		t.Error(err)
	}

	gipc.Sleep()

	cc, err2 := gipc.StartClient(NewClientConfig("test1010"))
	if err2 != nil {
		t.Error(err)
	}
	defer cc.Close()

	holdIt := make(chan bool, 1)

	go func() {
		for {
			m, _ := cc.Read()

			if m.Status == "Reconnecting" {
				holdIt <- false
				return
			}
		}
	}()

	for {

		mm, err2 := sc.Read()

		if err2 == nil {
			if mm.Status == "Connected" {
				sc.Close()
			}

			if mm.Status == "Closed" {
				break
			}
		}
	}

	<-holdIt
}
