package cidr

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func IsInRange2(trial net.IP, lower net.IP, upper net.IP) bool {
	if bytes.Compare(trial, lower) >= 0 && bytes.Compare(trial, upper) <= 0 {
		return true
	}
	return false
}

func ParseIPRange(start, end string) ([]string, error) {
	ip1 := net.ParseIP(start)
	ip2 := net.ParseIP(end)

	if ip1 == nil || ip2 == nil {
		return []string{}, fmt.Errorf("ip range invalid: %v %v", start, end)
	}

	var ips []string
	for ip := ip1; IsInRange2(ip, ip1, ip2); inc(ip) {
		ips = append(ips, ip.String())
	}
	return ips, nil
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

func TryFixNetMask(ip, cidr string) (string, error) {
	if ip == "" {
		return "", fmt.Errorf("ip allocate fail")
	}
	i := strings.IndexByte(ip, '/')
	if i < 0 {
		a := strings.Split(cidr, "/")
		if len(a) != 2 {
			return ip, fmt.Errorf("cidr parse error: %v", cidr)
		}
		return ip + "/" + a[1], nil
	}
	return ip, nil
}

func Hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		if len(ip) == 4 {
			if ip[3] == 0x00 || ip[3] == 0xff {
				continue
			}
		}
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address

	return ips, nil
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

func AllocateCIDR(cidr string, used []string) (string, error) {

	hosts, err := Hosts(cidr)
	if err != nil {
		return "", err
	}

	return AllocateInHosts(cidr, hosts, used)
}

func AllocateInHosts(cidr string, hosts []string, used []string) (string, error) {

	ips := diff(hosts, used)
	if len(ips) == 0 {
		return "", fmt.Errorf("no more useable cidr in %s", cidr)
	}

	// i := rand.Intn(len(ips))
	a := strings.Split(cidr, "/")
	return ips[0] + "/" + a[1], nil
}

func diff(slice1 []string, slice2 []string) []string {
	var diff []string
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			if !found {
				diff = append(diff, s1)
			}
		}
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}

func CalcDefaultGatewayByCIDR(cidr string) (string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}

	ip = ip.Mask(ipnet.Mask)
	inc(ip)
	return ip.String(), nil
}
