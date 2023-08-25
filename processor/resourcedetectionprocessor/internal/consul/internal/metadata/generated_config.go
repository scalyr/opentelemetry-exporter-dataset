// Code generated by mdatagen. DO NOT EDIT.

package metadata

// ResourceAttributeConfig provides common config for a particular resource attribute.
type ResourceAttributeConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// ResourceAttributesConfig provides config for resourcedetectionprocessor/consul resource attributes.
type ResourceAttributesConfig struct {
	AzureResourcegroupName ResourceAttributeConfig `mapstructure:"azure.resourcegroup.name"`
	AzureVMName            ResourceAttributeConfig `mapstructure:"azure.vm.name"`
	AzureVMScalesetName    ResourceAttributeConfig `mapstructure:"azure.vm.scaleset.name"`
	AzureVMSize            ResourceAttributeConfig `mapstructure:"azure.vm.size"`
	CloudAccountID         ResourceAttributeConfig `mapstructure:"cloud.account.id"`
	CloudPlatform          ResourceAttributeConfig `mapstructure:"cloud.platform"`
	CloudProvider          ResourceAttributeConfig `mapstructure:"cloud.provider"`
	CloudRegion            ResourceAttributeConfig `mapstructure:"cloud.region"`
	HostID                 ResourceAttributeConfig `mapstructure:"host.id"`
	HostName               ResourceAttributeConfig `mapstructure:"host.name"`
}

func DefaultResourceAttributesConfig() ResourceAttributesConfig {
	return ResourceAttributesConfig{
		AzureResourcegroupName: ResourceAttributeConfig{
			Enabled: true,
		},
		AzureVMName: ResourceAttributeConfig{
			Enabled: true,
		},
		AzureVMScalesetName: ResourceAttributeConfig{
			Enabled: true,
		},
		AzureVMSize: ResourceAttributeConfig{
			Enabled: true,
		},
		CloudAccountID: ResourceAttributeConfig{
			Enabled: true,
		},
		CloudPlatform: ResourceAttributeConfig{
			Enabled: true,
		},
		CloudProvider: ResourceAttributeConfig{
			Enabled: true,
		},
		CloudRegion: ResourceAttributeConfig{
			Enabled: true,
		},
		HostID: ResourceAttributeConfig{
			Enabled: true,
		},
		HostName: ResourceAttributeConfig{
			Enabled: true,
		},
	}
}
