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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	prismctfv1 "github.com/pwnlentoni/prism-ctf/api/v1"
	"github.com/pwnlentoni/prism-ctf/internal/utils"
)

var _ = Describe("ChallengeInstance Controller", func() {
	Context("When reconciling a resource", func() {
		const (
			challengeName = "test-isolated"
			resourceName  = "test-instance"
			team          = "testteam"
			randomID      = "abcd"
		)

		ctx := context.Background()

		// IsolatedChallenge and ChallengeInstance are cluster-scoped
		typeNamespacedName := types.NamespacedName{Name: resourceName}

		var controllerReconciler *ChallengeInstanceReconciler

		createInstance := func(expiration *metav1.Time) {
			By("creating the custom resource for the Kind ChallengeInstance")
			resource := &prismctfv1.ChallengeInstance{
				ObjectMeta: metav1.ObjectMeta{Name: resourceName},
				Spec: prismctfv1.ChallengeInstanceSpec{
					Team:       team,
					Challenge:  challengeName,
					Expiration: expiration,
					RandomId:   randomID,
					Flag:       "flag{test}",
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
		}

		reconcileInstance := func() (reconcile.Result, error) {
			return controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
		}

		BeforeEach(func() {
			controllerReconciler = &ChallengeInstanceReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: events.NewFakeRecorder(100),
			}

			By("creating the custom resource for the Kind IsolatedChallenge")
			err := k8sClient.Get(ctx, types.NamespacedName{Name: challengeName}, &prismctfv1.IsolatedChallenge{})
			if err != nil && errors.IsNotFound(err) {
				containers, exposes := testContainers()
				chal := &prismctfv1.IsolatedChallenge{
					ObjectMeta: metav1.ObjectMeta{Name: challengeName},
					Spec: prismctfv1.IsolatedChallengeSpec{
						Containers:   containers,
						Exposes:      exposes,
						Lifetime:     &metav1.Duration{Duration: 10 * time.Minute},
						FlagTemplate: "flag{test}",
					},
				}
				Expect(k8sClient.Create(ctx, chal)).To(Succeed())
			}
		})

		AfterEach(func() {
			By("Cleanup the ChallengeInstance, if it still exists")
			resource := &prismctfv1.ChallengeInstance{ObjectMeta: metav1.ObjectMeta{Name: resourceName}}
			Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, resource))).To(Succeed())
			Eventually(func() bool {
				return errors.IsNotFound(k8sClient.Get(ctx, typeNamespacedName, &prismctfv1.ChallengeInstance{}))
			}, 10*time.Second, 100*time.Millisecond).Should(BeTrue())
		})

		It("should reconcile an instance without expiration and never expire it", func() {
			createInstance(nil)

			By("Reconciling the created resource")
			result, err := reconcileInstance()
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeZero(), "no expiry requeue without an expiration")

			By("checking the instance still exists and is Ready")
			instance := &prismctfv1.ChallengeInstance{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, instance)).To(Succeed())
			Expect(instance.DeletionTimestamp).To(BeNil())
			cond := meta.FindStatusCondition(instance.Status.Conditions, "Ready")
			Expect(cond).NotTo(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue), cond.Message)
			Expect(cond.Reason).To(Equal("ReconcileSuccess"))

			By("checking the instance namespace and deployment")
			namespace := utils.IsolatedChallengeNamespace(challengeName, team, randomID)
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: namespace}, &corev1.Namespace{})).To(Succeed())
			deployment := &appsv1.Deployment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: testContainerName}, deployment)).To(Succeed())
			Expect(metav1.IsControlledBy(deployment, instance)).To(BeTrue())
		})

		It("should requeue an instance with a future expiration until it expires", func() {
			expiration := metav1.NewTime(time.Now().Add(5 * time.Minute))
			createInstance(&expiration)

			result, err := reconcileInstance()
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeNumerically(">", 4*time.Minute))
			Expect(result.RequeueAfter).To(BeNumerically("<=", 5*time.Minute+time.Second))
			Expect(k8sClient.Get(ctx, typeNamespacedName, &prismctfv1.ChallengeInstance{})).To(Succeed())
		})

		It("should delete an instance whose expiration is in the past", func() {
			expiration := metav1.NewTime(time.Now().Add(-time.Minute))
			createInstance(&expiration)

			result, err := reconcileInstance()
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeZero())
			Eventually(func() bool {
				return errors.IsNotFound(k8sClient.Get(ctx, typeNamespacedName, &prismctfv1.ChallengeInstance{}))
			}, 10*time.Second, 100*time.Millisecond).Should(BeTrue())
		})
	})
})
