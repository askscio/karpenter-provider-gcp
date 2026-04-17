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
	"strings"

	"github.com/cloudpilot-ai/karpenter-provider-gcp/pkg/apis/v1alpha1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// machineFamiliesPersistentDiskUnsupported lists GCE machine-type name prefixes (first segment of
// the machine type, e.g. "m4" for "m4-hypermem-16") whose VMs do not support Persistent Disk boot
// volumes and require Hyperdisk only. Keep aligned with GCP Compute Engine machine family docs.
//
// See: https://cloud.google.com/compute/docs/memory-optimized-machines (M4, X4 storage sections).
var machineFamiliesPersistentDiskUnsupported = sets.NewString("m4", "x4")

func bootDiskCategory(nodeClass *v1alpha1.GCENodeClass) string {
	if nodeClass == nil {
		return ""
	}
	for i := range nodeClass.Spec.Disks {
		d := &nodeClass.Spec.Disks[i]
		if d.Boot {
			return string(d.Category)
		}
	}
	return ""
}

func instanceTypeFamilyPrefix(machineTypeName string) string {
	parts := strings.Split(machineTypeName, "-")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func isPersistentDiskCategory(category string) bool {
	switch category {
	case "pd-standard", "pd-balanced", "pd-ssd", "pd-extreme":
		return true
	default:
		return false
	}
}

// instanceTypeCompatibleWithBootDisk returns false when the NodeClass boot disk category cannot be
// used with the given GCE machine type (so the provider should not offer that InstanceType).
func instanceTypeCompatibleWithBootDisk(nodeClass *v1alpha1.GCENodeClass, machineTypeName string) bool {
	cat := bootDiskCategory(nodeClass)
	if cat == "" {
		return true
	}
	family := instanceTypeFamilyPrefix(machineTypeName)
	if !machineFamiliesPersistentDiskUnsupported.Has(family) {
		return true
	}
	if isPersistentDiskCategory(cat) {
		return false
	}
	return strings.HasPrefix(cat, "hyperdisk")
}
