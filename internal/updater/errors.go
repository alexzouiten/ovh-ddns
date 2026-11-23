package updater

import "errors"

var (
	ErrNoV4Address = errors.New("host has no IPv4 address")
	ErrNoV6Address = errors.New("host has no IPv6 address")
)
