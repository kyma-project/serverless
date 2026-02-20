package fips

import (
	"crypto/fips140"
	"os"
)

func IsFIPS140Only() bool {
	return fips140.Enabled() && os.Getenv("GODEBUG") == "fips140=only,tlsmlkem=0"
}
