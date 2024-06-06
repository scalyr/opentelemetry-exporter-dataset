// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"go.opentelemetry.io/collector/confmap"
)

// ResourceAttributeConfig provides common config for a particular resource attribute.
type ResourceAttributeConfig struct {
	Enabled bool `mapstructure:"enabled"`

	enabledSetByUser bool
}

func (rac *ResourceAttributeConfig) Unmarshal(parser *confmap.Conf) error {
	if parser == nil {
		return nil
	}
	err := parser.Unmarshal(rac)
	if err != nil {
		return err
	}
	rac.enabledSetByUser = parser.IsSet("enabled")
	return nil
}

// ResourceAttributesConfig provides config for resourcedetectionprocessor/openshift resource attributes.
type ResourceAttributesConfig struct {
	CloudPlatform  ResourceAttributeConfig `mapstructure:"cloud.platform"`
	CloudProvider  ResourceAttributeConfig `mapstructure:"cloud.provider"`
	CloudRegion    ResourceAttributeConfig `mapstructure:"cloud.region"`
	K8sClusterName ResourceAttributeConfig `mapstructure:"k8s.cluster.name"`
}

func DefaultResourceAttributesConfig() ResourceAttributesConfig {
	return ResourceAttributesConfig{
		CloudPlatform: ResourceAttributeConfig{
			Enabled: true,
		},
		CloudProvider: ResourceAttributeConfig{
			Enabled: true,
		},
		CloudRegion: ResourceAttributeConfig{
			Enabled: true,
		},
		K8sClusterName: ResourceAttributeConfig{
			Enabled: true,
		},
	}
}
