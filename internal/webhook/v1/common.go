package v1

import (
	"fmt"
	challengesv1 "github.com/pwnlentoni/prism-ctf/api/v1"
)

func validateContainers(containers []challengesv1.ContainerSpec) (map[string]challengesv1.ContainerSpec, error) {
	m := make(map[string]challengesv1.ContainerSpec, len(containers))
	for _, spec := range containers {
		name := spec.Spec.Name
		ports := make(map[string]any, len(spec.Ports))
		for _, port := range spec.Ports {
			key := fmt.Sprint(port.Port, port.Protocol)
			if _, ok := ports[key]; ok {
				return nil, fmt.Errorf("duplicate port `%d` in container `%s`", port.Port, name)
			}
			ports[key] = nil
		}
		if _, ok := m[name]; ok {
			return nil, fmt.Errorf("duplicate container name `%s`", name)
		}
		m[name] = spec
	}
	return m, nil
}

func validateExposures(containers map[string]challengesv1.ContainerSpec, exposures []challengesv1.ExposeSpec) error {
	names := make(map[string]int)
	foundEmpty := false
	for _, exposure := range exposures {
		if len(exposure.Name) == 0 {
			if foundEmpty {
				return fmt.Errorf("challenge has multiple exposed ports without name")
			} else {
				foundEmpty = true
			}
		}
		if _, ok := names[exposure.Name]; ok {
			return fmt.Errorf("challenge has duplicated port name `%s`", exposure.Name)
		}
		container, ok := containers[exposure.Container]
		if !ok {
			return fmt.Errorf("port `%d` references non existing container `%s`", exposure.Port, exposure.Container)
		}
		found := false
		for _, port := range container.Ports {
			if exposure.Port == port.Port {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("port `%d` not exposed by container `%s`", exposure.Port, exposure.Container)
		}
	}
	return nil
}
