package dns

import (
	"testing"

	"github.com/alexzouiten/ovh-ddns/internal/utils"
)

func TestAString(t *testing.T) {
	actual := RecordA.String()
	if actual != "A" {
		t.Errorf("expected A record to stringify as 'A', got %s", actual)
	}
}

func TestAAAAString(t *testing.T) {
	actual := RecordAAAA.String()
	if actual != "AAAA" {
		t.Errorf("expected AAAA record to stringify as 'AAAA', got %s", actual)
	}
}

func TestUnknownRecordString(t *testing.T) {
	rt := RecordType(-1)
	utils.AssertPanic(t, func() { _ = rt.String() })
}

func TestAParse(t *testing.T) {
	var rt RecordType
	if err := rt.parse("A"); err != nil {
		t.Errorf("expected no error, got %s", err)
	}

	if rt != RecordA {
		t.Errorf("expected A record, got %s", rt)
	}
}

func TestAAAAParse(t *testing.T) {
	var rt RecordType
	if err := rt.parse("AAAA"); err != nil {
		t.Errorf("expected no error, got %s", err)
	}

	if rt != RecordAAAA {
		t.Errorf("expected AAAA record, got %s", rt)
	}
}

func TestUnknownRecordParse(t *testing.T) {
	var rt RecordType
	if err := rt.parse("test"); err == nil {
		t.Errorf("expected error, got none")
	}
}
