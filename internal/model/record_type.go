package model

import (
	"fmt"
	"strconv"
	"strings"
)

type RecordType int

const (
	RecordA RecordType = iota
	RecordAAAA
)

func RecordTypeFromString(s string) (RecordType, error) {
	switch strings.ToUpper(s) {
	case "A":
		return RecordA, nil
	case "AAAA":
		return RecordAAAA, nil
	default:
		return 0, fmt.Errorf("unknown record type '%s'", s)
	}
}

func (t RecordType) String() string {
	switch t {
	case RecordA:
		return "A"
	case RecordAAAA:
		return "AAAA"
	default:
		panic(fmt.Errorf("unsupported dns record type '%d'", t))
	}
}

func (t RecordType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *RecordType) UnmarshalText(data []byte) error {
	parsed, err := RecordTypeFromString(string(data))
	*t = parsed
	return err
}

func (t RecordType) MarshalJSON() ([]byte, error) {
	return strconv.AppendQuote(nil, t.String()), nil
}

func (t *RecordType) UnmarshalJSON(data []byte) error {
	s, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	parsed, err := RecordTypeFromString(s)
	*t = parsed
	return err
}
