//go:build network && randomize_ports

package gipc

import (
	"math/rand"
	"sync"
)

var mutex = &sync.Mutex{}
var ports = make(map[string]int, 200)

func randRange(min, max int) int {
	return rand.Intn(max-min) + min
}

func getPortByName(name string) int {
	mutex.Lock()
	port := ports[name]
	mutex.Unlock()
	return port
}

func setPortByName(name string, port int) {
	mutex.Lock()
	ports[name] = port
	mutex.Unlock()
}

func GetPort(name string) int {
	port := getPortByName(name)
	if port > 0 {
		return port
	} else {
		port = GetDefaultPort()
		port = randRange(port, port+200)
		setPortByName(name, port)
		return port
	}
}
