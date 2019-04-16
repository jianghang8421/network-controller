package cidr

import (
	"math/rand"
	"net"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

//  http://play.golang.org/p/m8TNTtygK0
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func Hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

func RandomIP(cidr string) (string, error) {
	ips, err := Hosts(cidr)
	if err != nil || len(ips) == 0 {
		return "", err
	}
	i := rand.Intn(len(ips))
	return ips[i], nil
}

func RandomCIDR(cidr string) (string, error) {
	ips, err := Hosts(cidr)
	if err != nil || len(ips) == 0 {
		return "", err
	}
	i := rand.Intn(len(ips))
	a := strings.Split(cidr, "/")
	return ips[i] + "/" + a[1], nil
}
