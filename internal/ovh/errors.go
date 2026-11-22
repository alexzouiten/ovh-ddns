package ovh

import "errors"

var (
	ErrMultipleRecords = errors.New("multiple records found for this type in the dns zone")
)
