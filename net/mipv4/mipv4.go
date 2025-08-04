package mipv4

import (
	"net"
)

var (
	// localIPv4 cache local IPv4 address.
	localIPv4 = ""
	// intranetIPv4Array cache intranet IPv4 address array.
	intranetIPv4Array []string
)

// GetIntranetIPArray gets the local intranet IPv4 address array.
func GetIntranetIPArray() ([]string, error) {
	if intranetIPv4Array != nil {
		return intranetIPv4Array, nil
	}
	ips := make([]string, 0)
	// get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return ips, err
	}
	// iterate all network interfaces
	for _, i := range interfaces {
		// skip down interfaces
		if i.Flags&net.FlagUp == 0 {
			continue
		}
		// get interface addresses
		addrs, err := i.Addrs()
		if err != nil {
			return ips, err
		}
		// iterate all addresses
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				// only process IPv4 addresses
				if ipv4 := ipNet.IP.To4(); ipv4 != nil {
					ip := ipv4.String()
					if IsIntranet(ip) {
						ips = append(ips, ip)
					}
				}
			}
		}
	}
	intranetIPv4Array = ips
	return ips, nil
}

// GetIPArray gets the local all IPv4 address array.
func GetIPArray() ([]string, error) {
	ips := make([]string, 0)
	interfaces, err := net.Interfaces()
	if err != nil {
		return ips, err
	}
	for _, i := range interfaces {
		if i.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			return ips, err
		}
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipv4 := ipNet.IP.To4(); ipv4 != nil {
					ips = append(ips, ipv4.String())
				}
			}
		}
	}
	return ips, nil
}

// GetLocalIP gets the local first IPv4 address.
func GetLocalIP() (string, error) {
	if localIPv4 != "" {
		return localIPv4, nil
	}
	ips, err := GetIntranetIPArray()
	if err != nil {
		return "", err
	}
	if len(ips) > 0 {
		localIPv4 = ips[0]
		return localIPv4, nil
	}
	return "", nil
}

// IsIntranet checks if the given IP is an intranet IP.
func IsIntranet(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.IsPrivate()
}
