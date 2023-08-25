// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package host

import (
	"context"
	"errors"
	"testing"
	"time"

	awsec2metadata "github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/awstesting/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockMetadataClient struct {
	count int
}

func (m *mockMetadataClient) GetInstanceIdentityDocumentWithContext(_ context.Context) (awsec2metadata.EC2InstanceIdentityDocument, error) {
	m.count++
	if m.count == 1 {
		return awsec2metadata.EC2InstanceIdentityDocument{}, errors.New("error")
	}

	return awsec2metadata.EC2InstanceIdentityDocument{
		Region:       "us-west-2",
		InstanceID:   "i-abcd1234",
		InstanceType: "c4.xlarge",
		PrivateIP:    "79.168.255.0",
	}, nil
}

func TestEC2Metadata(t *testing.T) {
	ctx := context.Background()
	sess := mock.Session
	instanceIDReadyC := make(chan bool)
	instanceIPReadyP := make(chan bool)
	clientOption := func(e *ec2Metadata) {
		e.client = &mockMetadataClient{}
	}
	e := newEC2Metadata(ctx, sess, 3*time.Millisecond, instanceIDReadyC, instanceIPReadyP, zap.NewNop(), clientOption)
	assert.NotNil(t, e)

	<-instanceIDReadyC
	<-instanceIPReadyP
	assert.Equal(t, "i-abcd1234", e.getInstanceID())
	assert.Equal(t, "c4.xlarge", e.getInstanceType())
	assert.Equal(t, "us-west-2", e.getRegion())
	assert.Equal(t, "79.168.255.0", e.getInstanceIP())
}
