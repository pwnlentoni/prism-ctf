package v1

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pwnlentoni/prism-ctf/internal/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateObject(client client.Client, l logr.Logger, expectedNs string, obj unstructured.Unstructured) error {
	hasNs, err := client.IsObjectNamespaced(&obj)
	if err != nil {
		return fmt.Errorf("is object namespaced: %w", err)
	}
	if hasNs && obj.GetNamespace() != expectedNs {
		return fmt.Errorf("object `%s/%s` of kind `%s` doesn't have the expected namespace `%s`", obj.GetNamespace(), obj.GetName(), obj.GetKind(), expectedNs)
	}
	return nil
}

func validateDoc(ctx context.Context, client client.Client, l logr.Logger, doc, expectedNs string) error {
	objs, err := utils.GetObjectsFromTemplate(client, ctx, doc)
	if err != nil {
		return err
	}

	for _, obj := range objs {
		err = validateObject(client, l, expectedNs, obj)
		if err != nil {
			l.Error(err, "object validation failed", "kind", obj.GetKind(), "name", obj.GetNamespace()+"/"+obj.GetName())
			return err
		}
		l.Info("got object from template", "kind", obj.GetKind(), "name", obj.GetNamespace()+"/"+obj.GetName())
	}
	return nil
}
