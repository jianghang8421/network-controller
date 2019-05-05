package cidr_test

import (
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
