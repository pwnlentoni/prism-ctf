package v1

import (
	"fmt"
	challengesv1 "github.com/pwnlentoni/prism-ctf/api/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func validateContainers(containers []challengesv1.ContainerSpec, path *field.Path) (map[string]challengesv1.ContainerSpec, error) {
	m := make(map[string]challengesv1.ContainerSpec, len(containers))
	for i, spec := range containers {
		path := path.Index(i)
		name := spec.Spec.Name
		ports := make(map[string]any, len(spec.Ports))
		for i, port := range spec.Ports {
			key := fmt.Sprint(port.Port, port.Protocol)
			if _, ok := ports[key]; ok {
				return nil, field.Duplicate(path.Child("ports").Index(i), port.Port)
			}
			ports[key] = nil
		}
		if _, ok := m[name]; ok {
			return nil, field.Duplicate(path.Child("spec", "name"), name)
		}
		m[name] = spec
	}
	return m, nil
}

func validateExposures(containers map[string]challengesv1.ContainerSpec, exposures []challengesv1.ExposeSpec, path *field.Path) error {
	names := make(map[string]int)
	foundEmpty := false
	for i, exposure := range exposures {
		path := path.Index(i)
		if len(exposure.Name) == 0 {
			if foundEmpty {
				return field.Invalid(path.Child("invalid"), exposure, "challenge has multiple exposed ports without name")
			} else {
				foundEmpty = true
			}
		}
		if _, ok := names[exposure.Name]; ok {
			return field.Duplicate(path.Child("name"), exposure.Name)
		}
		container, ok := containers[exposure.Container]
		if !ok {
			return field.NotFound(path.Child("container"), exposure.Container)
		}
		found := false
		for _, port := range container.Ports {
			if exposure.Port == port.Port {
				found = true
				break
			}
		}
		if !found {
			return field.Invalid(path.Child("port"), exposure.Port, "Port not exposed by container")
		}
	}
	return nil
}
