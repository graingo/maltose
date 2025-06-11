package mipv4

import (
	"net"
	"strings"
)

var (
	// localIPv4 cache local IPv4 address.
	localIPv4 = ""
	// intranetIPv4Array cache intranet IPv4 address array.
	intranetIPv4Array []string
)

// GetIntranetIpArray gets the local intranet IPv4 address array.
func GetIntranetIpArray() ([]string, error) {
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

// GetIpArray gets the local all IPv4 address array.
func GetIpArray() ([]string, error) {
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

// GetLocalIp gets the local first IPv4 address.
func GetLocalIp() (string, error) {
	if localIPv4 != "" {
		return localIPv4, nil
	}
	ips, err := GetIntranetIpArray()
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
	// split IP address
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	// check if it is an intranet IP
	// A range: 10.0.0.0--10.255.255.255
	// B range: 172.16.0.0--172.31.255.255
	// C range: 192.168.0.0--192.168.255.255
	if parts[0] == "10" ||
		(parts[0] == "172" && inRange(parts[1], 16, 31)) ||
		(parts[0] == "192" && parts[1] == "168") {
		return true
	}
	return false
}

// inRange checks if the string-represented number is within the specified range.
func inRange(s string, min, max int) bool {
	num := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
		num = num*10 + int(c-'0')
	}
	return num >= min && num <= max
}
