package fips

import (
	"crypto/fips140"
	"os"
)

const (
	GODEBUG_VALUE = "fips140=only,tlsmlkem=0"
)

type FipsChecker func() bool

func IsFIPS140Only() bool {
	return fips140.Enabled() && os.Getenv("GODEBUG") == GODEBUG_VALUE
}
