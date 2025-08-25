package config

const (
	// DefaultPort is the port under which
	// the REST API is exposed by default.
	DefaultPort = 6400
)

var (
	// buildFlavor determines the flavor of k8s-snap being built.
	// It can be overridden at build time using -ldflags "-X ...".
	buildFlavor = "default"
)

type Flavor string

const (
	FlavorDefault Flavor = "default"
	FlavorFIPS    Flavor = "fips"
)

// GetFlavor returns the flavor of k8s-snap being built.
// This function panics if the flavor set at build-time is unknown.
func GetFlavor() Flavor {
	flavor := Flavor(buildFlavor)

	switch flavor {
	case FlavorDefault, FlavorFIPS:
		return flavor
	default:
		panic("unknown build flavor: " + string(flavor))
	}
}
