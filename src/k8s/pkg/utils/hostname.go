package utils

import (
	"golang.org/x/net/idna"
)

// CleanHostname sanitises hostnames
var CleanHostname = idna.Lookup.ToASCII
