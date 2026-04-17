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

package instancetype

import (
	"testing"

	"github.com/cloudpilot-ai/karpenter-provider-gcp/pkg/apis/v1alpha1"
)

func TestInstanceTypeCompatibleWithBootDisk(t *testing.T) {
	t.Parallel()
	pdBalanced := &v1alpha1.GCENodeClass{
		Spec: v1alpha1.GCENodeClassSpec{
			Disks: []v1alpha1.Disk{
				{Boot: true, Category: "pd-balanced", SizeGiB: 100},
			},
		},
	}
	hyperdiskBoot := &v1alpha1.GCENodeClass{
		Spec: v1alpha1.GCENodeClassSpec{
			Disks: []v1alpha1.Disk{
				{Boot: true, Category: "hyperdisk-balanced", SizeGiB: 100},
			},
		},
	}
	tests := []struct {
		name   string
		nc     *v1alpha1.GCENodeClass
		mt     string
		wantOK bool
	}{
		{"m4 with pd-balanced", pdBalanced, "m4-hypermem-16", false},
		{"m4 with hyperdisk", hyperdiskBoot, "m4-hypermem-16", true},
		{"x4 with pd-balanced", pdBalanced, "x4-megamem-96", false},
		{"n2 with pd-balanced", pdBalanced, "n2-standard-16", true},
		{"m3 with pd-balanced", pdBalanced, "m3-megamem-128", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := instanceTypeCompatibleWithBootDisk(tt.nc, tt.mt)
			if got != tt.wantOK {
				t.Fatalf("instanceTypeCompatibleWithBootDisk() = %v, want %v", got, tt.wantOK)
			}
		})
	}
}
