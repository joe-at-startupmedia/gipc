//go:build network

package gipc

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func GetDefaultPort() int {
	envVar := os.Getenv("GIPC_NETWORK_PORT")
	if len(envVar) > 0 {
		valInt, err := strconv.Atoi(envVar)
		if err == nil {
			return valInt
		}
	}
	return DEFAULT_NETWORK_PORT
}

func GetDefaultHost() string {
	envVar := os.Getenv("GIPC_NETWORK_HOST")
	if len(envVar) > 0 {
		return envVar
	}
	return DEFAULT_NETWORK_HOST
}

func GetDefaultNetworkType() string {
	envVar := os.Getenv("GIPC_NETWORK_TYPE")
	if envVar == "udp" {
		return envVar
	}
	return DEFAULT_NETWORK_TYPE
}

func (c *Client) getHostAddr(clientId int) string {

	port := GetPort(c.config.ClientConfig.Name)
	return fmt.Sprintf("%s:%d", GetDefaultHost(), port+clientId)
}

func (s *Server) getHostAddr(clientId int) string {

	port := GetPort(s.config.ServerConfig.Name)
	return fmt.Sprintf("%s:%d", GetDefaultHost(), port+clientId)
}

func (c *Client) connect() (ActorConn, error) {

	conn, err := net.Dial(DEFAULT_NETWORK_TYPE, c.getHostAddr(c.ClientId))
	if err != nil {
		c.logger.Errorf("Dial error: %s", err)
	}

	return conn, err
}

func (s *Server) listen(clientId int) error {

	listener, err := net.Listen(DEFAULT_NETWORK_TYPE, s.getHostAddr(clientId))
	if err != nil {
		return err
	}

	s.listener = listener

	return nil
}

func (s *Server) accept() (ActorConn, error) {
	return s.listener.Accept()
}
