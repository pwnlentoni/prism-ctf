package utils

import (
	"context"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

var emptyTemplateError = errors.New("no objects in template")

func GetObjectsFromTemplate(client client.Client, ctx context.Context, doc string) ([]unstructured.Unstructured, error) {
	l := log.FromContext(ctx)
	objs := make([]unstructured.Unstructured, 0, 1)
	yDec := yaml.NewDecoder(strings.NewReader(doc))
	for {
		obj := &unstructured.Unstructured{Object: map[string]interface{}{}}
		err := yDec.Decode(obj.Object)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			l.Error(err, "failed to decode yaml")
			return nil, fmt.Errorf("yaml decode: %w", err)
		}

		if len(obj.Object) == 0 {
			l.Info("skipping empty obj")
			continue
		}

		gvk, err := client.GroupVersionKindFor(obj)
		if err != nil {
			l.Error(err, "failed to get gvk kind for object", "object", obj.Object)
			return nil, err
		}
		_, err = client.RESTMapper().RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			l.Error(err, "mapping error", "object", obj.Object)
			return nil, err
		}

		objs = append(objs, *obj)
	}

	if len(objs) == 0 {
		return nil, emptyTemplateError
	}
	return objs, nil
}
