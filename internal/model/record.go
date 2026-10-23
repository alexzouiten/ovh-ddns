package model

import "net"

type Record struct {
	Id        int
	Type      RecordType
	SubDomain string
	Target    net.IP
	Ttl       int
	Zone      string
}
