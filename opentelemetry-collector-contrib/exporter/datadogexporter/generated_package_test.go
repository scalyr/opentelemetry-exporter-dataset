// Code generated by mdatagen. DO NOT EDIT.

package datadogexporter

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	setupTestMain(m)
	// skipping goleak test as per metadata.yml configuration
	os.Exit(m.Run())
}
