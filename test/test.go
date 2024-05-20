package test

import (
	"github.com/joe-at-startupmedia/gipc"
	"os"
	"strconv"
	"time"
)

func GetTimeout() time.Duration {
	envVar := os.Getenv("GIPC_TIMEOUT")
	if len(envVar) > 0 {
		valInt, err := strconv.Atoi(envVar)
		if err == nil {
			return time.Duration(valInt)
		}
	}
	return 3
}

var TimeoutClientConfig = gipc.ClientConfig{
	Timeout:    time.Second * GetTimeout(),
	Encryption: gipc.ENCRYPT_BY_DEFAULT,
}

func NewClientTimeoutConfig(name string) *gipc.ClientConfig {
	config := TimeoutClientConfig
	config.Name = name
	return &config
}

func NewServerConfig(name string) *gipc.ServerConfig {
	return &gipc.ServerConfig{Name: name, Encryption: gipc.ENCRYPT_BY_DEFAULT}
}

func NewClientConfig(name string) *gipc.ClientConfig {
	return &gipc.ClientConfig{Name: name, Encryption: gipc.ENCRYPT_BY_DEFAULT}
}
