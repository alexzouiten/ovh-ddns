package model

import (
	"fmt"
	"net"
	"testing"
)

func TestZoneStringer(t *testing.T) {
	ipv4 := "12.34.56.78"
	ipv6 := "fd12:34:56::1"
	zone := Zone{
		A: &Record{
			Target: net.ParseIP(ipv4),
		},
		AAAA: &Record{
			Target: net.ParseIP(ipv6),
		},
	}

	expected := fmt.Sprintf("{A: %s, AAAA: %s}", ipv4, ipv6)
	actual := zone.String()
	if actual != expected {
		t.Errorf("expected string '%s', got '%s'", expected, actual)
	}
}

func TestZoneStringer_NilIpv4(t *testing.T) {
	ipv4 := "12.34.56.78"
	zone := Zone{
		A: &Record{
			Target: net.ParseIP(ipv4),
		},
		AAAA: nil,
	}

	expected := fmt.Sprintf("{A: %s, AAAA: nil}", ipv4)
	actual := zone.String()
	if actual != expected {
		t.Errorf("expected string '%s', got '%s'", expected, actual)
	}
}

func TestZoneStringer_NilIpv6(t *testing.T) {
	ipv6 := "fd12:34:56::1"
	zone := Zone{
		A: nil,
		AAAA: &Record{
			Target: net.ParseIP(ipv6),
		},
	}

	expected := fmt.Sprintf("{A: nil, AAAA: %s}", ipv6)
	actual := zone.String()
	if actual != expected {
		t.Errorf("expected string '%s', got '%s'", expected, actual)
	}
}
