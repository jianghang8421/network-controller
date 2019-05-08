package cidr_test

import (
	"reflect"
	"testing"

	"github.com/cnrancher/network-controller/cidr"
)

func Test_Hosts(t *testing.T) {
	subnet := "10.244.0.0/16"
	ips, err := cidr.Hosts(subnet)
	if err != nil {
		t.Error(err)
	}
	t.Log(len(ips))
}

func Test_AllocateCIDR(t *testing.T) {
	subnet := "192.168.1.0/24"
	ips, err := cidr.Hosts(subnet)
	if err != nil {
		t.Error(err)
	}

	if len(ips) != 253 {
		t.Error(len(ips))
	}

	ip, err := cidr.AllocateCIDR(subnet, ips[1:])
	if ip != "192.168.1.2/24" {
		t.Error(ip)
	}
}

func Test_CalcGatewayByCIDR(t *testing.T) {
	subnet := "192.168.56.0/24"
	ip, err := cidr.CalcGatewayByCIDR(subnet)
	if err != nil {
		t.Error(err)
	}

	if ip != "192.168.56.1" {
		t.Log(ip)
	}
}

func Test_ParseIPRange(t *testing.T) {
	ip1 := "192.168.1.100"
	ip2 := "192.168.1.105"
	ips, err := cidr.ParseIPRange(ip1, ip2)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(ips, []string{"192.168.1.100", "192.168.1.101", "192.168.1.102", "192.168.1.103", "192.168.1.104", "192.168.1.105"}) {
		t.Error(ips)
	}
}
