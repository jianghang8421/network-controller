package controller

import (
	"crypto/sha1"
	"fmt"
	"testing"
)

func Test_hash(t *testing.T) {
	hash := sha1.Sum([]byte("192.168.56.1-123.345.14.144"))
	r := fmt.Sprintf("%x", hash)
	t.Log(r)
}
