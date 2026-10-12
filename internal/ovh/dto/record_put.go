package dto

import (
	"net"

	"github.com/alexzouiten/ovh-ddns/internal/model"
)

type RecordPut struct {
	SubDomain string `json:"subDomain,omitempty"`
	Target    net.IP `json:"target"`
	Ttl       int    `json:"ttl,omitempty"`
}

func NewRecordPutDtoFromModel(r model.Record) RecordPut {
	return RecordPut{
		SubDomain: r.SubDomain, // optional
		Target:    r.Target,
		Ttl:       r.Ttl, // optional
	}
}
