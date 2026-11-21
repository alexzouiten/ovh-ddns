package ovh

import "errors"

var (
	ErrNoRecord        = errors.New("no record of this type in the dns zone")
	ErrMultipleRecords = errors.New("multiple records found for this type in the dns zone")
)
