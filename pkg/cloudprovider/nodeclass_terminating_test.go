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

package cloudprovider

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"

	"github.com/cloudpilot-ai/karpenter-provider-gcp/pkg/apis/v1alpha1"
)

// terminatingNodeClassClient is a minimal client.Client that serves a
// GCENodeClass carrying a DeletionTimestamp and a NodePool referencing it.
// It models the production trigger: a NodePool still pointing at a
// GCENodeClass that is being re-applied/churned by a release rollout.
type terminatingNodeClassClient struct {
	client.Client
}

func (c *terminatingNodeClassClient) Get(_ context.Context, key client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
	switch o := obj.(type) {
	case *v1alpha1.GCENodeClass:
		o.Name = key.Name
		o.DeletionTimestamp = ptr(metav1.Now())
		o.Finalizers = []string{"karpenter.k8s.gcp/termination"}
	case *karpv1.NodePool:
		o.Name = key.Name
		o.Spec.Template.Spec.NodeClassRef = &karpv1.NodeClassReference{Name: "generic-nodepool-100"}
	}
	return nil
}

// TestResolveNodeClassFromNodePoolTerminating asserts the sentinel error is
// returned instead of the (nil, nil) pair that previously let a nil
// *GCENodeClass reach callers.
func TestResolveNodeClassFromNodePoolTerminating(t *testing.T) {
	c := &CloudProvider{kubeClient: &terminatingNodeClassClient{}}

	nodeClass, err := c.resolveNodeClassFromNodePool(context.Background(), &karpv1.NodePool{
		Spec: karpv1.NodePoolSpec{
			Template: karpv1.NodeClaimTemplate{
				Spec: karpv1.NodeClaimTemplateSpec{
					NodeClassRef: &karpv1.NodeClassReference{Name: "generic-nodepool-100"},
				},
			},
		},
	})

	require.True(t, stderrors.Is(err, errNodeClassTerminating))
	require.Nil(t, nodeClass)
}

// TestGetInstanceTypesTerminatingNodeClassNoPanic reproduces the crash at
// pkg/providers/instancetype/instancetype.go:138, where a nil nodeClass was
// dereferenced via nodeClass.Spec.KubeletConfiguration. The guard must return
// early so the nil never reaches the instance type provider.
func TestGetInstanceTypesTerminatingNodeClassNoPanic(t *testing.T) {
	c := &CloudProvider{kubeClient: &terminatingNodeClassClient{}}

	instanceTypes, err := c.GetInstanceTypes(context.Background(), &karpv1.NodePool{
		Spec: karpv1.NodePoolSpec{
			Template: karpv1.NodeClaimTemplate{
				Spec: karpv1.NodeClaimTemplateSpec{
					NodeClassRef: &karpv1.NodeClassReference{Name: "generic-nodepool-100"},
				},
			},
		},
	})

	require.NoError(t, err)
	require.Empty(t, instanceTypes)
}

// TestIsDriftedTerminatingNodeClassNoPanic reproduces the crash at
// pkg/cloudprovider/drift.go:59, where areStaticFieldsDrifted dereferenced
// nodeClass.Annotations on a nil nodeClass. This is the path the
// nodeclaim.disruption (Drift) controller took in the grammarly outage.
func TestIsDriftedTerminatingNodeClassNoPanic(t *testing.T) {
	c := &CloudProvider{kubeClient: &terminatingNodeClassClient{}}

	nc := &karpv1.NodeClaim{}
	nc.Labels = map[string]string{karpv1.NodePoolLabelKey: "generic-nodepool-100"}

	reason, err := c.IsDrifted(context.Background(), nc)

	require.NoError(t, err)
	require.Empty(t, reason)
}
