package dto

import (
	"net"

	"github.com/alexzouiten/ovh-ddns/internal/model"
)

type RecordPost struct {
	FieldType string `json:"fieldType"`
	SubDomain string `json:"subDomain,omitempty"`
	Target    net.IP `json:"target"`
	Ttl       int    `json:"ttl,omitempty"`
}

func NewRecordPostDtoFromModel(r model.Record) RecordPost {
	return RecordPost{
		FieldType: r.RecordType.String(),
		SubDomain: r.SubDomain, // optional
		Target:    r.Target,
		Ttl:       r.Ttl, // optional
	}
}
