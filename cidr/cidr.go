package cidr

import (
	"fmt"
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
	return ips[2 : len(ips)-1], nil
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

	ips := diff(hosts, used)
	if len(ips) == 0 {
		return "", fmt.Errorf("no more useable cidr in %s", cidr)
	}

	i := rand.Intn(len(ips))
	a := strings.Split(cidr, "/")
	return ips[i] + "/" + a[1], nil
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

func CalcGatewayByCIDR(cidr string) (string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}

	ip = ip.Mask(ipnet.Mask)
	inc(ip)
	return ip.String(), nil
}
