// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package fileconsumer

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/featuregate"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer/matcher"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/operatortest"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/parser/regex"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/testutil"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/tokenize"
)

func TestUnmarshal(t *testing.T) {
	operatortest.ConfigUnmarshalTests{
		DefaultConfig: newMockOperatorConfig(NewConfig()),
		TestsFile:     filepath.Join(".", "testdata", "config.yaml"),
		Tests: []operatortest.ConfigUnmarshalTest{
			{
				Name: "include_one",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "one.log")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "include_multi",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "one.log", "two.log", "three.log")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "include_glob",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "*.log")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "include_glob_double_asterisk",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "**.log")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "include_glob_double_asterisk_nested",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "directory/**/*.log")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "include_glob_double_asterisk_prefix",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "**/directory/**/*.log")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "include_inline",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "a.log", "b.log")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "include_string",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "aString")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "exclude_one",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "*.log")
					cfg.Exclude = append(cfg.Exclude, "one.log")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "exclude_multi",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "*.log")
					cfg.Exclude = append(cfg.Exclude, "one.log", "two.log", "three.log")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "exclude_glob",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "*.log")
					cfg.Exclude = append(cfg.Exclude, "not*.log")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "exclude_glob_double_asterisk",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "*.log")
					cfg.Exclude = append(cfg.Exclude, "not**.log")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "exclude_glob_double_asterisk_nested",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "*.log")
					cfg.Exclude = append(cfg.Exclude, "directory/**/not*.log")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "exclude_glob_double_asterisk_prefix",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "*.log")
					cfg.Exclude = append(cfg.Exclude, "**/directory/**/not*.log")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "exclude_inline",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "*.log")
					cfg.Exclude = append(cfg.Exclude, "a.log", "b.log")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "exclude_string",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Include = append(cfg.Include, "*.log")
					cfg.Exclude = append(cfg.Exclude, "aString")
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "sort_by_timestamp",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.OrderingCriteria = matcher.OrderingCriteria{
						Regex: `err\.[a-zA-Z]\.\d+\.(?P<rotation_time>\d{10})\.log`,
						SortBy: []matcher.Sort{
							{
								SortType:  "timestamp",
								RegexKey:  "rotation_time",
								Ascending: true,
								Location:  "utc",
								Layout:    `%Y%m%d%H`,
							},
						},
					}
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "sort_by_numeric",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.OrderingCriteria = matcher.OrderingCriteria{
						Regex: `err\.(?P<file_num>[a-zA-Z])\.\d+\.\d{10}\.log`,
						SortBy: []matcher.Sort{
							{
								SortType: "numeric",
								RegexKey: "file_num",
							},
						},
					}
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "poll_interval_no_units",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.PollInterval = time.Second
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "poll_interval_1s",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.PollInterval = time.Second
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "poll_interval_1ms",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.PollInterval = time.Millisecond
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "poll_interval_1000ms",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.PollInterval = time.Second
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "fingerprint_size_no_units",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.FingerprintSize = helper.ByteSize(1000)
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "fingerprint_size_1kb_lower",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.FingerprintSize = helper.ByteSize(1000)
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "fingerprint_size_1KB",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.FingerprintSize = helper.ByteSize(1000)
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "fingerprint_size_1kib_lower",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.FingerprintSize = helper.ByteSize(1024)
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "fingerprint_size_1KiB",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.FingerprintSize = helper.ByteSize(1024)
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "fingerprint_size_float",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.FingerprintSize = helper.ByteSize(1100)
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "multiline_line_start_string",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					newSplit := tokenize.NewSplitterConfig()
					newSplit.Multiline.LineStartPattern = "Start"
					cfg.Splitter = newSplit
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "multiline_line_start_special",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					newSplit := tokenize.NewSplitterConfig()
					newSplit.Multiline.LineStartPattern = "%"
					cfg.Splitter = newSplit
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "multiline_line_end_string",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					newSplit := tokenize.NewSplitterConfig()
					newSplit.Multiline.LineEndPattern = "Start"
					cfg.Splitter = newSplit
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "multiline_line_end_special",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					newSplit := tokenize.NewSplitterConfig()
					newSplit.Multiline.LineEndPattern = "%"
					cfg.Splitter = newSplit
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "start_at_string",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.StartAt = "beginning"
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "max_concurrent_large",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.MaxConcurrentFiles = 9223372036854775807
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "max_log_size_mib_lower",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.MaxLogSize = helper.ByteSize(1048576)
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "max_log_size_mib_upper",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.MaxLogSize = helper.ByteSize(1048576)
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "max_log_size_mb_upper",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.MaxLogSize = helper.ByteSize(1048576)
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "max_log_size_mb_lower",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.MaxLogSize = helper.ByteSize(1048576)
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "encoding_lower",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Splitter.Encoding = "utf-16le"
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "encoding_upper",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.Splitter.Encoding = "UTF-16lE"
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "max_batches_1",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					cfg.MaxBatches = 1
					return newMockOperatorConfig(cfg)
				}(),
			},
			{
				Name: "header_config",
				Expect: func() *mockOperatorConfig {
					cfg := NewConfig()
					regexCfg := regex.NewConfig()
					cfg.Header = &HeaderConfig{
						Pattern: "^#",
						MetadataOperators: []operator.Config{
							{
								Builder: regexCfg,
							},
						},
					}
					return newMockOperatorConfig(cfg)
				}(),
			},
		},
	}.Run(t)
}

func TestBuild(t *testing.T) {
	t.Parallel()

	basicConfig := func() *Config {
		cfg := NewConfig()
		cfg.Include = []string{"/var/log/testpath.*"}
		cfg.Exclude = []string{"/var/log/testpath.ex*"}
		cfg.PollInterval = 10 * time.Millisecond
		return cfg
	}

	cases := []struct {
		name             string
		modifyBaseConfig func(*Config)
		errorRequirement require.ErrorAssertionFunc
		validate         func(*testing.T, *Manager)
	}{
		{
			"Basic",
			func(f *Config) {},
			require.NoError,
			func(t *testing.T, f *Manager) {
				require.Equal(t, f.pollInterval, 10*time.Millisecond)
			},
		},
		{
			"BadIncludeGlob",
			func(f *Config) {
				f.Include = []string{"["}
			},
			require.Error,
			nil,
		},
		{
			"BadExcludeGlob",
			func(f *Config) {
				f.Include = []string{"["}
			},
			require.Error,
			nil,
		},
		{
			"MultilineConfiguredStartAndEndPatterns",
			func(f *Config) {
				f.Splitter = tokenize.NewSplitterConfig()
				f.Splitter.Multiline = tokenize.MultilineConfig{
					LineEndPattern:   "Exists",
					LineStartPattern: "Exists",
				}
			},
			require.Error,
			nil,
		},
		{
			"MultilineConfiguredStartPattern",
			func(f *Config) {
				f.Splitter = tokenize.NewSplitterConfig()
				f.Splitter.Multiline = tokenize.MultilineConfig{
					LineStartPattern: "START.*",
				}
			},
			require.NoError,
			func(t *testing.T, f *Manager) {},
		},
		{
			"MultilineConfiguredEndPattern",
			func(f *Config) {
				f.Splitter = tokenize.NewSplitterConfig()
				f.Splitter.Multiline = tokenize.MultilineConfig{
					LineEndPattern: "END.*",
				}
			},
			require.NoError,
			func(t *testing.T, f *Manager) {},
		},
		{
			"InvalidEncoding",
			func(f *Config) {
				f.Splitter.Encoding = "UTF-3233"
			},
			require.Error,
			nil,
		},
		{
			"LineStartAndEnd",
			func(f *Config) {
				f.Splitter = tokenize.NewSplitterConfig()
				f.Splitter.Multiline = tokenize.MultilineConfig{
					LineStartPattern: ".*",
					LineEndPattern:   ".*",
				}
			},
			require.Error,
			nil,
		},
		{
			"NoLineStartOrEnd",
			func(f *Config) {
				f.Splitter = tokenize.NewSplitterConfig()
				f.Splitter.Multiline = tokenize.MultilineConfig{}
			},
			require.NoError,
			func(t *testing.T, f *Manager) {},
		},
		{
			"InvalidLineStartRegex",
			func(f *Config) {
				f.Splitter = tokenize.NewSplitterConfig()
				f.Splitter.Multiline = tokenize.MultilineConfig{
					LineStartPattern: "(",
				}
			},
			require.Error,
			nil,
		},
		{
			"InvalidLineEndRegex",
			func(f *Config) {
				f.Splitter = tokenize.NewSplitterConfig()
				f.Splitter.Multiline = tokenize.MultilineConfig{
					LineEndPattern: "(",
				}
			},
			require.Error,
			nil,
		},
		{
			"InvalidStartAtDelete",
			func(f *Config) {
				f.StartAt = "end"
				f.DeleteAfterRead = true
			},
			require.Error,
			nil,
		},
		{
			"InvalidMaxBatches",
			func(f *Config) {
				f.MaxBatches = -1
			},
			require.Error,
			nil,
		},
		{
			"ValidMaxBatches",
			func(f *Config) {
				f.MaxBatches = 6
			},
			require.NoError,
			func(t *testing.T, m *Manager) {
				require.Equal(t, 6, m.maxBatches)
			},
		},
		{
			"HeaderConfigNoFlag",
			func(f *Config) {
				f.Header = &HeaderConfig{}
			},
			require.Error,
			nil,
		},
		{
			"BadOrderingCriteriaRegex",
			func(f *Config) {
				f.OrderingCriteria = matcher.OrderingCriteria{
					SortBy: []matcher.Sort{
						{
							SortType: "numeric",
							RegexKey: "value",
						},
					},
				}
			},
			require.Error,
			nil,
		},
		{
			"OrderingCriteriaTimestampMissingLayout",
			func(f *Config) {
				f.OrderingCriteria = matcher.OrderingCriteria{
					Regex: ".*",
					SortBy: []matcher.Sort{
						{
							SortType: "timestamp",
							RegexKey: "value",
						},
					},
				}
			},
			require.Error,
			nil,
		},
		{
			"GoodOrderingCriteriaTimestamp",
			func(f *Config) {
				f.OrderingCriteria = matcher.OrderingCriteria{
					Regex: ".*",
					SortBy: []matcher.Sort{
						{
							SortType: "timestamp",
							RegexKey: "value",
							Layout:   "%Y%m%d%H",
						},
					},
				}
			},
			require.NoError,
			func(t *testing.T, f *Manager) {},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()
			cfg := basicConfig()
			tc.modifyBaseConfig(cfg)

			input, err := cfg.Build(testutil.Logger(t), nopEmitFunc)
			tc.errorRequirement(t, err)
			if err != nil {
				return
			}

			tc.validate(t, input)
		})
	}
}

func TestBuildWithSplitFunc(t *testing.T) {
	t.Parallel()

	basicConfig := func() *Config {
		cfg := NewConfig()
		cfg.Include = []string{"/var/log/testpath.*"}
		cfg.Exclude = []string{"/var/log/testpath.ex*"}
		cfg.PollInterval = 10 * time.Millisecond
		return cfg
	}

	cases := []struct {
		name             string
		modifyBaseConfig func(*Config)
		errorRequirement require.ErrorAssertionFunc
		validate         func(*testing.T, *Manager)
	}{
		{
			"Basic",
			func(f *Config) {},
			require.NoError,
			func(t *testing.T, f *Manager) {
				require.Equal(t, f.pollInterval, 10*time.Millisecond)
			},
		},
		{
			"BadIncludeGlob",
			func(f *Config) {
				f.Include = []string{"["}
			},
			require.Error,
			nil,
		},
		{
			"BadExcludeGlob",
			func(f *Config) {
				f.Include = []string{"["}
			},
			require.Error,
			nil,
		},
		{
			"InvalidEncoding",
			func(f *Config) {
				f.Splitter.Encoding = "UTF-3233"
			},
			require.Error,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()
			cfg := basicConfig()
			tc.modifyBaseConfig(cfg)

			splitNone := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
				if !atEOF {
					return 0, nil, nil
				}
				if len(data) == 0 {
					return 0, nil, nil
				}
				return len(data), data, nil
			}

			input, err := cfg.BuildWithSplitFunc(testutil.Logger(t), nopEmitFunc, splitNone)
			tc.errorRequirement(t, err)
			if err != nil {
				return
			}

			tc.validate(t, input)
		})
	}
}

func TestBuildWithHeader(t *testing.T) {
	require.NoError(t, featuregate.GlobalRegistry().Set(AllowHeaderMetadataParsing.ID(), true))
	t.Cleanup(func() {
		require.NoError(t, featuregate.GlobalRegistry().Set(AllowHeaderMetadataParsing.ID(), false))
	})

	basicConfig := func() *Config {
		cfg := NewConfig()
		cfg.Include = []string{"/var/log/testpath.*"}
		cfg.Exclude = []string{"/var/log/testpath.ex*"}
		cfg.PollInterval = 10 * time.Millisecond
		return cfg
	}

	cases := []struct {
		name             string
		modifyBaseConfig func(*Config)
		errorRequirement require.ErrorAssertionFunc
		validate         func(*testing.T, *Manager)
	}{
		{
			"InvalidHeaderConfig",
			func(f *Config) {
				f.Header = &HeaderConfig{}
				f.StartAt = "beginning"
			},
			require.Error,
			nil,
		},
		{
			"HeaderConfigWithStartAtEnd",
			func(f *Config) {
				regexCfg := regex.NewConfig()
				regexCfg.Regex = "^(?P<field>.*)"
				f.Header = &HeaderConfig{
					Pattern: "^#",
					MetadataOperators: []operator.Config{
						{
							Builder: regexCfg,
						},
					},
				}
				f.StartAt = "end"
			},
			require.Error,
			nil,
		},
		{
			"ValidHeaderConfig",
			func(f *Config) {
				regexCfg := regex.NewConfig()
				regexCfg.Regex = "^(?P<field>.*)"
				f.Header = &HeaderConfig{
					Pattern: "^#",
					MetadataOperators: []operator.Config{
						{
							Builder: regexCfg,
						},
					},
				}
				f.StartAt = "beginning"
			},
			require.NoError,
			func(t *testing.T, f *Manager) {
				require.NotNil(t, f.readerFactory.headerConfig.SplitFunc)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()
			cfg := basicConfig()
			tc.modifyBaseConfig(cfg)

			input, err := cfg.Build(testutil.Logger(t), nopEmitFunc)
			tc.errorRequirement(t, err)
			if err != nil {
				return
			}
			tc.validate(t, input)
		})
	}
}
