package model

import (
	"net"
	"testing"

	"github.com/alexzouiten/ovh-ddns/internal/utils"
)

func TestIpsEqual(t *testing.T) {
	a := Ips{
		V4: net.ParseIP("1.2.3.4"),
		V6: net.ParseIP("2001:db8::68"),
	}
	b := a

	c := Ips{
		V4: net.ParseIP("4.3.2.1"),
		V6: net.ParseIP("2002:db8::68"),
	}

	if !a.Equal(b) {
		t.Fatal("IPs should be equal")
	}

	if a.Equal(c) {
		t.Fatal("IPs should not be equal")
	}
}

func TestIpsEqual_Nil(t *testing.T) {
	a := Ips{
		V4: nil,
		V6: nil,
	}
	b := Ips{
		V4: net.ParseIP("1.2.3.4"),
		V6: net.ParseIP("2001:db8::45"),
	}

	utils.AssertNoPanic(t, func() { a.Equal(b) })
}

func TestForRecordType(t *testing.T) {
	ips := Ips{
		V4: net.ParseIP("123.123.123.123"),
		V6: net.ParseIP("2001:db8::4"),
	}

	if !ips.ForRecordType(RecordA).Equal(ips.V4) {
		t.Errorf("expected IPv4 to be returned")
	}

	if !ips.ForRecordType(RecordAAAA).Equal(ips.V6) {
		t.Errorf("expected IPv6 to be returned")
	}
}
