package test

import (
	"github.com/joe-at-startupmedia/gipc"
	"time"
)

var TimeoutClientConfig = gipc.ClientConfig{
	Timeout:    time.Second * 3,
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
