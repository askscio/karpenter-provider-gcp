/*
Copyright 2025 The CloudPilot AI Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metadata

import (
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/compute/v1"
)

func TestAppendImageStreamingInjectsIntoKubeEnv(t *testing.T) {
	md := &compute.Metadata{Items: []*compute.MetadataItems{
		{Key: "kube-env", Value: swag.String("CLUSTER_NAME: c")},
		{Key: "kube-labels", Value: swag.String("job-type=generic")},
	}}

	AppendImageStreaming(md)

	require.Equal(t, "CLUSTER_NAME: c\nENABLE_GCFS: \"true\"", swag.StringValue(md.Items[0].Value))
	require.Equal(t, "job-type=generic", swag.StringValue(md.Items[1].Value), "non kube-env metadata untouched")
}

func TestAppendImageStreamingIsIdempotent(t *testing.T) {
	md := &compute.Metadata{Items: []*compute.MetadataItems{
		{Key: "kube-env", Value: swag.String("ENABLE_GCFS: \"true\"")},
	}}

	AppendImageStreaming(md)

	require.Equal(t, "ENABLE_GCFS: \"true\"", swag.StringValue(md.Items[0].Value))
}
