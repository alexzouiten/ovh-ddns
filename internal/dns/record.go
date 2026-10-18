package dns

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
	return t.parse(string(data))
}

func (t RecordType) MarshalJSON() ([]byte, error) {
	return strconv.AppendQuote(nil, t.String()), nil
}

func (t *RecordType) UnmarshalJSON(data []byte) error {
	s, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	return t.parse(s)
}

func (t *RecordType) parse(s string) error {
	switch strings.ToUpper(s) {
	case "A":
		*t = RecordA
	case "AAAA":
		*t = RecordAAAA
	default:
		return fmt.Errorf("unknown record type '%s'", s)
	}
	return nil
}
