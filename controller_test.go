package main

import (
	"reflect"
	"testing"

	v1 "github.com/cnrancher/network-controller/types/apis/macvlan/v1"
)

func Test_RemoveDuplicatesFromSlice(t *testing.T) {
	a := []string{"a", "b", "b", "c", "a", "c"}
	b := RemoveDuplicatesFromSlice(a)
	if !reflect.DeepEqual(b, []string{"a", "b", "c"}) {
		t.Error(b)
	}
}

func Test_CalcHostsFromRanges(t *testing.T) {
	hosts := CalcHostsFromRanges([]v1.IPRange{
		v1.IPRange{"192.168.1.2", "192.168.1.5"},
		v1.IPRange{"192.168.1.3", "192.168.1.6"},
		v1.IPRange{"192.168.1.9", "192.168.1.11"}})

	if !reflect.DeepEqual(hosts, []string{
		"192.168.1.2",
		"192.168.1.3",
		"192.168.1.4",
		"192.168.1.5",
		"192.168.1.6",
		"192.168.1.9",
		"192.168.1.10",
		"192.168.1.11",
	}) {
		t.Error(hosts)
	}
}
