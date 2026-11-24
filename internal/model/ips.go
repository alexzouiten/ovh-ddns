package model

import (
	"fmt"
	"net"
)

type Ips struct {
	V4 net.IP
	V6 net.IP
}

func (i *Ips) Equal(other Ips) bool {
	return i.V4.Equal(other.V4) && i.V6.Equal(other.V6)
}

func (i *Ips) ForRecordType(recordType RecordType) net.IP {
	switch recordType {
	case RecordA:
		return i.V4
	case RecordAAAA:
		return i.V6
	default:
		panic(fmt.Errorf("unsupported record type"))
	}
}
