package fips

import (
	"crypto/fips140"
	"fmt"
	"os"
	"strings"
)

const (
	FIPS140OnlyEnvVar = "fips140=only"
	TLSMLKEMEnvVar    = "tlsmlkem=0"
)

var (
	GODEBUG_VALUE = fmt.Sprintf("%s,%s", FIPS140OnlyEnvVar, TLSMLKEMEnvVar)
)

type FipsChecker func() bool

func IsFIPS140Only() bool {
	godebug := os.Getenv("GODEBUG")
	return fips140.Enabled() && strings.Contains(godebug, FIPS140OnlyEnvVar) && strings.Contains(godebug, TLSMLKEMEnvVar)
}
