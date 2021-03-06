package controller

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net"
	"strings"

	"github.com/cnrancher/network-controller/ipcalc"
	macvlanv1 "github.com/cnrancher/network-controller/types/apis/macvlan/v1"
	v1 "github.com/cnrancher/network-controller/types/apis/macvlan/v1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) allocateAutoIP(pod *corev1.Pod, subnet *macvlanv1.MacvlanSubnet) (net.IP, string, error) {
	ips, err := c.macvlanClientset.MacvlanV1().
		MacvlanIPs("").
		List(metav1.ListOptions{LabelSelector: "subnet=" + subnet.Name})

	// add gateway ip to used list
	used := append(convertToIPList(ips), net.ParseIP(subnet.Spec.Gateway))
	hosts, err := calcSubnetHosts(subnet)
	if err != nil {
		c.eventMacvlanSubnetError(pod, err)
	}

	usable := ipcalc.RemoveUsedHosts(hosts, used)
	if len(usable) == 0 {
		return nil, "", fmt.Errorf(messageNoEnoughIP, subnet.Name)
	}
	ip, cidr := usable[0], convertIPtoCIDR(usable[0], subnet.Spec.CIDR)
	return ip, cidr, nil
}

func convertToIPList(ips *v1.MacvlanIPList) []net.IP {
	used := []net.IP{}
	for _, item := range ips.Items {
		ip := strings.Split(item.Spec.CIDR, "/")
		if len(ip) == 2 {
			used = append(used, net.ParseIP(ip[0]))
		}
	}
	return used
}

func convertIPtoCIDR(ip net.IP, cidr string) string {
	nets := strings.Split(cidr, "/")
	suffix := ""
	if len(nets) == 2 {
		suffix = nets[1]
	}
	return ip.String() + "/" + suffix
}

func calcSubnetHosts(subnet *macvlanv1.MacvlanSubnet) ([]net.IP, error) {
	hosts, err := ipcalc.CIDRtoHosts(subnet.Spec.CIDR)
	if err != nil {
		return nil, err
	}

	ranges := calcHostsInRanges(subnet.Spec.Ranges)
	if len(ranges) != 0 {
		usable := ipcalc.GetUsableHosts(hosts, ranges)
		return usable, nil
	}

	return hosts, nil
}

func calcHostsInRanges(ranges []v1.IPRange) []net.IP {
	hosts := []net.IP{}
	for _, v := range ranges {
		ips := ipcalc.ParseIPRange(v.RangeStart, v.RangeEnd)
		hosts = append(hosts, ips...)
	}
	return removeDuplicatesFromSlice(hosts)
}

func removeDuplicatesFromSlice(hosts []net.IP) []net.IP {
	m := make(map[string]bool)
	result := []net.IP{}
	for _, item := range hosts {
		if _, ok := m[item.String()]; ok {

		} else {
			m[item.String()] = true
			result = append(result, item)
		}
	}
	return result
}

func isSingleIP(ip string) bool {
	return nil != net.ParseIP(ip)
}

func (c *Controller) allocateSingleIP(pod *corev1.Pod, subnet *macvlanv1.MacvlanSubnet, ipString string) (net.IP, string, error) {
	ips, err := c.macvlanClientset.MacvlanV1().
		MacvlanIPs("").
		List(metav1.ListOptions{LabelSelector: "subnet=" + subnet.Name})

	// add gateway ip to used list
	used := append(convertToIPList(ips), net.ParseIP(subnet.Spec.Gateway))
	hosts, err := calcSubnetHosts(subnet)
	if err != nil {
		return nil, "", err
	}

	ip := net.ParseIP(ipString)

	if err := checkIPValidation(ip, hosts, used, subnet); err != nil {
		return nil, "", err
	}

	cidr := convertIPtoCIDR(ip, subnet.Spec.CIDR)
	return ip, cidr, nil
}

func checkIPValidation(ip net.IP, hosts, used []net.IP, subnet *macvlanv1.MacvlanSubnet) error {
	if !isInHosts(hosts, ip) {
		return fmt.Errorf("%s invalid in %s", ip.String(), subnet.Name)
	}

	if isInHosts(used, ip) {
		return fmt.Errorf("%s is used in %s", ip.String(), subnet.Name)
	}
	return nil
}

func isInHosts(h []net.IP, ip net.IP) bool {
	for _, v := range h {
		if bytes.Compare(v, ip) == 0 {
			return true
		}
	}
	return false
}

func isMultipleIP(ip string) bool {
	if !strings.Contains(ip, "-") {
		return false
	}
	ips := strings.Split(strings.Trim(ip, " "), "-")

	if len(ips) < 2 {
		return false
	}

	for _, v := range ips {
		if net.ParseIP(v) == nil {
			return false
		}
	}
	return true
}

func (c *Controller) allocateMultipleIP(pod *corev1.Pod, subnet *macvlanv1.MacvlanSubnet, annotationIP string, annotationMac string) (net.IP, string, string, error) {

	macs := []string{}
	ips := strings.Split(strings.Trim(annotationIP, " "), "-")

	if strings.Contains(annotationMac, "-") {
		macs = strings.Split(strings.Trim(annotationMac, " "), "-")
		if len(macs) != len(ips) {
			return nil, "", "", fmt.Errorf("count of multiple IP and Mac not same: %s %s", annotationIP, annotationMac)
		}
	}

	ipUnused := map[string]bool{}
	ipToMac := map[string]string{}

	for i, v := range ips {
		ipUnused[v] = true
		if len(macs) != 0 {
			ipToMac[v] = macs[i]
		}
	}

	hash := fmt.Sprintf("%x", sha1.Sum([]byte(annotationIP)))
	ret, err := c.kubeClientset.CoreV1().
		Pods(pod.Namespace).
		List(metav1.ListOptions{LabelSelector: macvlanv1.LabelMultipleIPHash + "=" + hash})

	if err != nil {
		return nil, "", "", err
	}

	log.Infof("labeled pod count： %v", len(ret.Items))
	for _, v := range ret.Items {
		labelIP := v.Labels[macvlanv1.LabelSelectedIP]
		if labelIP != "" && ipUnused[labelIP] == true {
			ipUnused[labelIP] = false
		}
	}

	for _, key := range ips {
		if ipUnused[key] {
			ip := net.ParseIP(key)

			ips, err := c.macvlanClientset.MacvlanV1().
				MacvlanIPs("").
				List(metav1.ListOptions{LabelSelector: "subnet=" + subnet.Name})

			// add gateway ip to used list
			used := append(convertToIPList(ips), net.ParseIP(subnet.Spec.Gateway))
			hosts, err := calcSubnetHosts(subnet)
			if err != nil {
				return nil, "", "", err
			}

			if err := checkIPValidation(ip, hosts, used, subnet); err != nil {
				return nil, "", "", err
			}

			mac := ""
			if len(macs) != 0 {
				mac = ipToMac[key]
			}

			return ip, convertIPtoCIDR(ip, subnet.Spec.CIDR), mac, nil
		}
	}

	// error event: no enough ip
	return nil, "", "", fmt.Errorf("No enough ip resouce in subnet: %s", annotationIP)
}
