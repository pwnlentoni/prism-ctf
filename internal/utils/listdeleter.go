package utils

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"slices"
)

type ListDeleter[T client.Object] struct {
	client client.Client
	items  []T
}

func NewDeleter[T client.Object](ctx context.Context, c client.Client, obj T, namespace string) (*ListDeleter[T], error) {
	ld := &ListDeleter[T]{
		client: c,
	}
	l := &unstructured.UnstructuredList{}

	gvk, err := c.GroupVersionKindFor(obj)
	if err != nil {
		return nil, fmt.Errorf("gvkfor: %w", err)
	}
	l.SetGroupVersionKind(gvk)
	err = ld.client.List(ctx, l, client.InNamespace(namespace))
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}
	ld.items = make([]T, len(l.Items))
	for i, item := range l.Items {
		var o T
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(item.UnstructuredContent(), &o)
		if err != nil {
			return nil, fmt.Errorf("from unstructured: %w", err)
		}
		ld.items[i] = o
	}
	return ld, nil
}

func (ld *ListDeleter[T]) MarkUsed(ctx context.Context, obj T) {
	if ld == nil {
		log.FromContext(ctx).Info("MarkUsed called on nil ListDeleter", "object", obj)
		return
	}
	ld.items = slices.DeleteFunc(ld.items, func(t T) bool {
		return t.GetNamespace() == obj.GetNamespace() && t.GetName() == obj.GetName()
	})
}

func (ld *ListDeleter[T]) DeleteUnused(ctx context.Context) error {
	l := log.FromContext(ctx)
	if ld == nil {
		l.Info("DeleteUnused called on nil ListDeleter")
		return nil
	}
	for _, item := range ld.items {
		err := ld.client.Delete(ctx, item)
		l.Info("deleting object", "kind", item.GetObjectKind().GroupVersionKind().String(), "name", item.GetNamespace()+item.GetName())
		if err != nil {
			return err
		}
	}
	return nil
}
