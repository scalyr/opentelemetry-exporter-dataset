// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ec2

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/awstesting/mock"
	"github.com/stretchr/testify/assert"
)

func TestMetadataProviderGetError(t *testing.T) {
	type args struct {
		ctx  context.Context
		sess *session.Session
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "mock session",
			args: args{
				ctx:  context.Background(),
				sess: mock.Session,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewProvider(tt.args.sess)
			_, err := c.Get(tt.args.ctx)
			assert.Error(t, err)
		})
	}
}

func TestMetadataProvider_available(t *testing.T) {
	type fields struct {
	}
	type args struct {
		ctx  context.Context
		sess *session.Session
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   error
	}{
		{
			name:   "mock session",
			fields: fields{},
			args:   args{ctx: context.Background(), sess: mock.Session},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewProvider(tt.args.sess)
			_, err := c.InstanceID(tt.args.ctx)
			assert.ErrorIs(t, err, tt.want)
		})
	}
}
