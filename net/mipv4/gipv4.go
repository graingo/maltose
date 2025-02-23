package mipv4

import (
	"net"
	"strings"
)

var (
	// localIPv4 缓存本地IPv4地址
	localIPv4 = ""
	// intranetIPv4Array 缓存内网IPv4地址数组
	intranetIPv4Array []string
)

// GetIntranetIpArray 获取本地内网IPv4地址数组
func GetIntranetIpArray() ([]string, error) {
	if intranetIPv4Array != nil {
		return intranetIPv4Array, nil
	}
	ips := make([]string, 0)
	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return ips, err
	}
	// 遍历所有网络接口
	for _, i := range interfaces {
		// 跳过 down 的接口
		if i.Flags&net.FlagUp == 0 {
			continue
		}
		// 获取接口的地址
		addrs, err := i.Addrs()
		if err != nil {
			return ips, err
		}
		// 遍历所有地址
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				// 只处理IPv4地址
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

// GetIpArray 获取本地所有IPv4地址数组
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

// GetLocalIp 获取本地首个IPv4地址
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

// IsIntranet 判断所给IP是否为内网IP
func IsIntranet(ip string) bool {
	// 分割IP地址
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	// 判断是否为内网IP
	// A类: 10.0.0.0--10.255.255.255
	// B类: 172.16.0.0--172.31.255.255
	// C类: 192.168.0.0--192.168.255.255
	if parts[0] == "10" ||
		(parts[0] == "172" && inRange(parts[1], 16, 31)) ||
		(parts[0] == "192" && parts[1] == "168") {
		return true
	}
	return false
}

// inRange 判断字符串表示的数字是否在指定范围内
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
