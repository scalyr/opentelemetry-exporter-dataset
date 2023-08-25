// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestMetricsBuilderConfig(t *testing.T) {
	tests := []struct {
		name string
		want MetricsBuilderConfig
	}{
		{
			name: "default",
			want: DefaultMetricsBuilderConfig(),
		},
		{
			name: "all_set",
			want: MetricsBuilderConfig{
				Metrics: MetricsConfig{
					FlinkJobCheckpointCount:           MetricConfig{Enabled: true},
					FlinkJobCheckpointInProgress:      MetricConfig{Enabled: true},
					FlinkJobLastCheckpointSize:        MetricConfig{Enabled: true},
					FlinkJobLastCheckpointTime:        MetricConfig{Enabled: true},
					FlinkJobRestartCount:              MetricConfig{Enabled: true},
					FlinkJvmClassLoaderClassesLoaded:  MetricConfig{Enabled: true},
					FlinkJvmCPULoad:                   MetricConfig{Enabled: true},
					FlinkJvmCPUTime:                   MetricConfig{Enabled: true},
					FlinkJvmGcCollectionsCount:        MetricConfig{Enabled: true},
					FlinkJvmGcCollectionsTime:         MetricConfig{Enabled: true},
					FlinkJvmMemoryDirectTotalCapacity: MetricConfig{Enabled: true},
					FlinkJvmMemoryDirectUsed:          MetricConfig{Enabled: true},
					FlinkJvmMemoryHeapCommitted:       MetricConfig{Enabled: true},
					FlinkJvmMemoryHeapMax:             MetricConfig{Enabled: true},
					FlinkJvmMemoryHeapUsed:            MetricConfig{Enabled: true},
					FlinkJvmMemoryMappedTotalCapacity: MetricConfig{Enabled: true},
					FlinkJvmMemoryMappedUsed:          MetricConfig{Enabled: true},
					FlinkJvmMemoryMetaspaceCommitted:  MetricConfig{Enabled: true},
					FlinkJvmMemoryMetaspaceMax:        MetricConfig{Enabled: true},
					FlinkJvmMemoryMetaspaceUsed:       MetricConfig{Enabled: true},
					FlinkJvmMemoryNonheapCommitted:    MetricConfig{Enabled: true},
					FlinkJvmMemoryNonheapMax:          MetricConfig{Enabled: true},
					FlinkJvmMemoryNonheapUsed:         MetricConfig{Enabled: true},
					FlinkJvmThreadsCount:              MetricConfig{Enabled: true},
					FlinkMemoryManagedTotal:           MetricConfig{Enabled: true},
					FlinkMemoryManagedUsed:            MetricConfig{Enabled: true},
					FlinkOperatorRecordCount:          MetricConfig{Enabled: true},
					FlinkOperatorWatermarkOutput:      MetricConfig{Enabled: true},
					FlinkTaskRecordCount:              MetricConfig{Enabled: true},
				},
				ResourceAttributes: ResourceAttributesConfig{
					FlinkJobName:       ResourceAttributeConfig{Enabled: true},
					FlinkResourceType:  ResourceAttributeConfig{Enabled: true},
					FlinkSubtaskIndex:  ResourceAttributeConfig{Enabled: true},
					FlinkTaskName:      ResourceAttributeConfig{Enabled: true},
					FlinkTaskmanagerID: ResourceAttributeConfig{Enabled: true},
					HostName:           ResourceAttributeConfig{Enabled: true},
				},
			},
		},
		{
			name: "none_set",
			want: MetricsBuilderConfig{
				Metrics: MetricsConfig{
					FlinkJobCheckpointCount:           MetricConfig{Enabled: false},
					FlinkJobCheckpointInProgress:      MetricConfig{Enabled: false},
					FlinkJobLastCheckpointSize:        MetricConfig{Enabled: false},
					FlinkJobLastCheckpointTime:        MetricConfig{Enabled: false},
					FlinkJobRestartCount:              MetricConfig{Enabled: false},
					FlinkJvmClassLoaderClassesLoaded:  MetricConfig{Enabled: false},
					FlinkJvmCPULoad:                   MetricConfig{Enabled: false},
					FlinkJvmCPUTime:                   MetricConfig{Enabled: false},
					FlinkJvmGcCollectionsCount:        MetricConfig{Enabled: false},
					FlinkJvmGcCollectionsTime:         MetricConfig{Enabled: false},
					FlinkJvmMemoryDirectTotalCapacity: MetricConfig{Enabled: false},
					FlinkJvmMemoryDirectUsed:          MetricConfig{Enabled: false},
					FlinkJvmMemoryHeapCommitted:       MetricConfig{Enabled: false},
					FlinkJvmMemoryHeapMax:             MetricConfig{Enabled: false},
					FlinkJvmMemoryHeapUsed:            MetricConfig{Enabled: false},
					FlinkJvmMemoryMappedTotalCapacity: MetricConfig{Enabled: false},
					FlinkJvmMemoryMappedUsed:          MetricConfig{Enabled: false},
					FlinkJvmMemoryMetaspaceCommitted:  MetricConfig{Enabled: false},
					FlinkJvmMemoryMetaspaceMax:        MetricConfig{Enabled: false},
					FlinkJvmMemoryMetaspaceUsed:       MetricConfig{Enabled: false},
					FlinkJvmMemoryNonheapCommitted:    MetricConfig{Enabled: false},
					FlinkJvmMemoryNonheapMax:          MetricConfig{Enabled: false},
					FlinkJvmMemoryNonheapUsed:         MetricConfig{Enabled: false},
					FlinkJvmThreadsCount:              MetricConfig{Enabled: false},
					FlinkMemoryManagedTotal:           MetricConfig{Enabled: false},
					FlinkMemoryManagedUsed:            MetricConfig{Enabled: false},
					FlinkOperatorRecordCount:          MetricConfig{Enabled: false},
					FlinkOperatorWatermarkOutput:      MetricConfig{Enabled: false},
					FlinkTaskRecordCount:              MetricConfig{Enabled: false},
				},
				ResourceAttributes: ResourceAttributesConfig{
					FlinkJobName:       ResourceAttributeConfig{Enabled: false},
					FlinkResourceType:  ResourceAttributeConfig{Enabled: false},
					FlinkSubtaskIndex:  ResourceAttributeConfig{Enabled: false},
					FlinkTaskName:      ResourceAttributeConfig{Enabled: false},
					FlinkTaskmanagerID: ResourceAttributeConfig{Enabled: false},
					HostName:           ResourceAttributeConfig{Enabled: false},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := loadMetricsBuilderConfig(t, tt.name)
			if diff := cmp.Diff(tt.want, cfg, cmpopts.IgnoreUnexported(MetricConfig{}, ResourceAttributeConfig{})); diff != "" {
				t.Errorf("Config mismatch (-expected +actual):\n%s", diff)
			}
		})
	}
}

func loadMetricsBuilderConfig(t *testing.T, name string) MetricsBuilderConfig {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)
	sub, err := cm.Sub(name)
	require.NoError(t, err)
	cfg := DefaultMetricsBuilderConfig()
	require.NoError(t, component.UnmarshalConfig(sub, &cfg))
	return cfg
}

func TestResourceAttributesConfig(t *testing.T) {
	tests := []struct {
		name string
		want ResourceAttributesConfig
	}{
		{
			name: "default",
			want: DefaultResourceAttributesConfig(),
		},
		{
			name: "all_set",
			want: ResourceAttributesConfig{
				FlinkJobName:       ResourceAttributeConfig{Enabled: true},
				FlinkResourceType:  ResourceAttributeConfig{Enabled: true},
				FlinkSubtaskIndex:  ResourceAttributeConfig{Enabled: true},
				FlinkTaskName:      ResourceAttributeConfig{Enabled: true},
				FlinkTaskmanagerID: ResourceAttributeConfig{Enabled: true},
				HostName:           ResourceAttributeConfig{Enabled: true},
			},
		},
		{
			name: "none_set",
			want: ResourceAttributesConfig{
				FlinkJobName:       ResourceAttributeConfig{Enabled: false},
				FlinkResourceType:  ResourceAttributeConfig{Enabled: false},
				FlinkSubtaskIndex:  ResourceAttributeConfig{Enabled: false},
				FlinkTaskName:      ResourceAttributeConfig{Enabled: false},
				FlinkTaskmanagerID: ResourceAttributeConfig{Enabled: false},
				HostName:           ResourceAttributeConfig{Enabled: false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := loadResourceAttributesConfig(t, tt.name)
			if diff := cmp.Diff(tt.want, cfg, cmpopts.IgnoreUnexported(ResourceAttributeConfig{})); diff != "" {
				t.Errorf("Config mismatch (-expected +actual):\n%s", diff)
			}
		})
	}
}

func loadResourceAttributesConfig(t *testing.T, name string) ResourceAttributesConfig {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)
	sub, err := cm.Sub(name)
	require.NoError(t, err)
	sub, err = sub.Sub("resource_attributes")
	require.NoError(t, err)
	cfg := DefaultResourceAttributesConfig()
	require.NoError(t, component.UnmarshalConfig(sub, &cfg))
	return cfg
}
