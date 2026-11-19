/*
Copyright 2026.

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

package controller

import (
	"context"

	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	prismctfv1 "github.com/pwnlentoni/prism-ctf/api/v1"
	"github.com/pwnlentoni/prism-ctf/internal/utils"
)

const (
	testContainerName  = "backend"
	testContainerImage = "traefik/whoamitcp:latest"
)

// testContainers returns a minimal valid container/expose spec, modeled after config/samples.
func testContainers() ([]prismctfv1.ContainerSpec, []prismctfv1.ExposeSpec) {
	return []prismctfv1.ContainerSpec{{
		Ports:    []prismctfv1.PortSpec{{Port: 8080, Protocol: corev1.ProtocolTCP}},
		Egress:   false,
		Replicas: 1,
		Spec: &corev1.Container{
			Name:  testContainerName,
			Image: testContainerImage,
		},
	}}, []prismctfv1.ExposeSpec{{
		Container: testContainerName,
		Port:      8080,
		Protocol:  prismctfv1.ExposeProtocolTCP,
	}}
}

var _ = Describe("SharedChallenge Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-shared"

		ctx := context.Background()

		// SharedChallenge is cluster-scoped
		typeNamespacedName := types.NamespacedName{Name: resourceName}

		BeforeEach(func() {
			By("creating the custom resource for the Kind SharedChallenge")
			err := k8sClient.Get(ctx, typeNamespacedName, &prismctfv1.SharedChallenge{})
			if err != nil && errors.IsNotFound(err) {
				containers, exposes := testContainers()
				resource := &prismctfv1.SharedChallenge{
					Name: resourceName,
					Spec: prismctfv1.SharedChallengeSpec{
						Containers: containers,
						Exposes:    exposes,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &prismctfv1.SharedChallenge{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance SharedChallenge")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &SharedChallengeReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: events.NewFakeRecorder(100),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("checking the Ready condition")
			chal := &prismctfv1.SharedChallenge{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, chal)).To(Succeed())
			cond := meta.FindStatusCondition(chal.Status.Conditions, "Ready")
			Expect(cond).NotTo(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue), cond.Message)
			Expect(cond.Reason).To(Equal("ReconcileSuccess"))
			Expect(chal.Status.ExposedUrls).To(HaveLen(1))
			Expect(chal.Status.ExposedUrls[0].Hostname).To(Equal(resourceName + utils.DomainSuffix()))

			namespace := utils.SharedChallengeNamespace(resourceName)

			By("checking the challenge namespace")
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: namespace}, &corev1.Namespace{})).To(Succeed())

			By("checking the container deployment")
			deployment := &appsv1.Deployment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: testContainerName}, deployment)).To(Succeed())
			Expect(deployment.Spec.Template.Spec.Containers).NotTo(BeEmpty())
			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal(testContainerImage))
			Expect(metav1.IsControlledBy(deployment, chal)).To(BeTrue())

			By("checking the cilium network policies")
			for _, name := range []string{"isolate", "allow-egress"} {
				Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name},
					&ciliumv2.CiliumNetworkPolicy{})).To(Succeed())
			}
		})
	})
})
