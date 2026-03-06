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

func doesContainValue(parts []string, want string) bool {
	for _, p := range parts {
		if strings.EqualFold(strings.TrimSpace(p), want) {
			return true
		}
	}
	return false
}

func IsFIPS140Only() bool {
	godebug := os.Getenv("GODEBUG")
	if !fips140.Enabled() {
		return false
	}

	parts := strings.Split(godebug, ",")
	return doesContainValue(parts, FIPS140OnlyEnvVar) && doesContainValue(parts, TLSMLKEMEnvVar)
}
