//go:build network && !randomize_ports

package gipc

func GetPort(_ string) int {
	return GetDefaultPort()
}
