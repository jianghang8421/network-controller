package controller

import (
	"fmt"

	"github.com/cnrancher/network-controller/ipcalc"
	macvlanv1 "github.com/cnrancher/network-controller/types/apis/macvlan/v1"
	log "github.com/sirupsen/logrus"
)

func (c *Controller) onMacvlanSubnetAdd(obj interface{}) {

	subnet, ok := obj.(*macvlanv1.MacvlanSubnet)
	if !ok {
		return
	}

	log.Errorf("MacvlanSubnets Add : %s %v", subnet.Name, subnet)

	if subnet.Labels == nil {
		subnet.Labels = map[string]string{}
	}
	subnet.Labels["master"] = subnet.Spec.Master
	subnet.Labels["vlan"] = fmt.Sprint(subnet.Spec.VLAN)
	subnet.Labels["mode"] = subnet.Spec.Mode

	if subnet.Spec.Gateway == "" {
		gateway, err := ipcalc.CalcDefaultGateway(subnet.Spec.CIDR)
		if err != nil {
			log.Errorf("CalcGatewayByCIDR error : %v %s", err, subnet.Spec.CIDR)
		}
		subnet.Spec.Gateway = gateway.String()
	}

	_, err := c.macvlanClientset.MacvlanV1().MacvlanSubnets("kube-system").Update(subnet)
	if err != nil {
		log.Errorf("MacvlanSubnets Update : %v %s %v", err, subnet.Name, subnet)
	}
}
