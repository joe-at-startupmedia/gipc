//go:build network && randomize_ports

package gipc

import (
	"math/rand"
)

var ports = make(map[string]int, 200)

func randRange(min, max int) int {
	return rand.Intn(max-min) + min
}

func GetPort(name string) int {
	var port int
	if ports[name] > 0 {
		port = ports[name]
	} else {
		port = GetDefaultPort()
		port = randRange(port, port+200)
		ports[name] = port
	}

	return port
}
