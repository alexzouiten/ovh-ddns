package config

import "errors"

var (
	ErrNoRecord               = errors.New("need at least one configured record")
	ErrAppKeyEmpty            = errors.New("app key cannot be empty")
	ErrAppSecretEmpty         = errors.New("app secret cannot be empty")
	ErrConsumerKeyEmpty       = errors.New("consumer key cannot be empty")
	ErrUpdateIntervalTooShort = errors.New("update interval cannot be shorter than 1min")
	ErrIpProviderUrlEmpty     = errors.New("ip provider url cannot be empty")
	ErrTtlNegative            = errors.New("dns record ttl cannot be negative")
	ErrRecordTypeDuplicated   = errors.New("record types must only appear once per record")
)
