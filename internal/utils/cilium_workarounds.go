package utils

import (
	"fmt"
	slimmetav1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/apis/meta/v1"
	ciliumapi "github.com/cilium/cilium/pkg/policy/api"
	"k8s.io/apimachinery/pkg/api/equality"
)

// cilium uses some structs with non exported fields which make reflect.DeepEqual panic
// the issue is fixed by providing custom comparison functions for them

func RegisterCiliumComparisons() error {
	err := equality.Semantic.AddFunc(func(a, b slimmetav1.Time) bool {
		return a.DeepEqual(&b)
	})
	if err != nil {
		return fmt.Errorf("slim_metav1.Time: %w", err)
	}
	err = equality.Semantic.AddFunc(func(a, b ciliumapi.EndpointSelector) bool {
		return equality.Semantic.DeepEqual(a.LabelSelector, b.LabelSelector)
	})
	if err != nil {
		return fmt.Errorf("ciliumapi.EndpointSelector: %w", err)
	}
	return nil
}
