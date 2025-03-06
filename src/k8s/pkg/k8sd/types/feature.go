package types

import (
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
)

type Image struct {
	Registry   string `yaml:"registry,omitempty"`
	Repository string `yaml:"repository,omitempty"`
	Tag        string `yaml:"tag,omitempty"`
}

func (i Image) GetURI() string {
	if i.Registry == "" {
		return i.Repository
	}

	return fmt.Sprintf("%s/%s", i.Registry, i.Repository)
}

type FeatureManifest struct {
	Name    string `yaml:"name,omitempty"`
	Version string `yaml:"version,omitempty"`

	Charts map[string]helm.InstallableChart `yaml:"charts,omitempty"`

	Images map[string]Image `yaml:"images,omitempty"`
}

func (f FeatureManifest) GetName() string {
	return f.Name
}

func (f FeatureManifest) GetVersion() string {
	return f.Version
}

func (f FeatureManifest) GetChart(name string) helm.InstallableChart {
	return f.Charts[name]
}

func (f FeatureManifest) GetImage(name string) Image {
	return f.Images[name]
}
