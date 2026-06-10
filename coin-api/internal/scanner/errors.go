package scanner

import "errors"

var (
	ErrNoConfig  = errors.New("no .coin/config.yaml")
	ErrConfigV1  = errors.New("config v1 not supported")
	ErrBadConfig = errors.New("invalid config v2")
)
