/*
Copyright 2025.

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

package v1

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/pwnlentoni/prism-ctf/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	prismctfv1 "github.com/pwnlentoni/prism-ctf/api/v1"
)

// nolint:unused
// log is for logging in this package.
var challengeinstancelog = logf.Log.WithName("challengeinstance-resource")

// SetupChallengeInstanceWebhookWithManager registers the webhook for ChallengeInstance in the manager.
func SetupChallengeInstanceWebhookWithManager(mgr ctrl.Manager) error {
	err := mgr.GetCache().IndexField(context.Background(), &prismctfv1.ChallengeInstance{}, ".spec.team", func(object client.Object) []string {
		i := object.(*prismctfv1.ChallengeInstance)
		return []string{i.Spec.Team}
	})
	if err != nil {
		return err
	}
	return ctrl.NewWebhookManagedBy(mgr).For(&prismctfv1.ChallengeInstance{}).
		WithValidator(&ChallengeInstanceCustomValidator{Client: mgr.GetClient()}).
		WithDefaulter(&ChallengeInstanceCustomDefaulter{Client: mgr.GetClient()}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-prism-ctf-pwnlentoni-team-v1-challengeinstance,mutating=true,failurePolicy=fail,sideEffects=None,groups=prism-ctf.pwnlentoni.team,resources=challengeinstances,verbs=create,versions=v1,name=mchallengeinstance-v1.kb.io,admissionReviewVersions=v1

// ChallengeInstanceCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind ChallengeInstance when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type ChallengeInstanceCustomDefaulter struct {
	Client client.Client
}

var _ webhook.CustomDefaulter = &ChallengeInstanceCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind ChallengeInstance.
func (d *ChallengeInstanceCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	i, ok := obj.(*prismctfv1.ChallengeInstance)

	if !ok {
		return fmt.Errorf("expected an ChallengeInstance object but got %T", obj)
	}
	challengeinstancelog.Info("Defaulting for ChallengeInstance", "name", i.GetName())

	if len(i.Spec.RandomId) != 0 {
		return errors.New("random id already set on new instance")
	}

	rid := make([]byte, *utils.RandomTokenLength)
	_, err := rand.Read(rid)
	if err != nil {
		challengeinstancelog.Error(err, "failed to generate instance id")
		return errors.New("failed to generate random id")
	}
	i.Spec.RandomId = hex.EncodeToString(rid)

	tpl := &prismctfv1.IsolatedChallenge{}
	err = d.Client.Get(ctx, client.ObjectKey{Name: i.Spec.Challenge}, tpl)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return field.NotFound(field.NewPath("spec", "challenge"), i.Spec.Challenge)
		} else {
			challengeinstancelog.Error(err, "error while getting challenge template")
			return errors.New("failed to get challenge")
		}
	}
	i.Spec.Expiration = &metav1.Time{Time: time.Now().Add(tpl.Spec.Lifetime.Duration)}

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-prism-ctf-pwnlentoni-team-v1-challengeinstance,mutating=false,failurePolicy=fail,sideEffects=None,groups=prism-ctf.pwnlentoni.team,resources=challengeinstances,verbs=create,versions=v1,name=vchallengeinstance-v1.kb.io,admissionReviewVersions=v1

// ChallengeInstanceCustomValidator struct is responsible for validating the ChallengeInstance resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ChallengeInstanceCustomValidator struct {
	Client client.Client
}

var _ webhook.CustomValidator = &ChallengeInstanceCustomValidator{}

var ErrLimitReached = errors.New("instance count limit reached")
var ErrAlreadyExists = errors.New("instance already exists")

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type ChallengeInstance.
func (v *ChallengeInstanceCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	i, ok := obj.(*prismctfv1.ChallengeInstance)
	if !ok {
		return nil, fmt.Errorf("expected a ChallengeInstance object but got %T", obj)
	}
	challengeinstancelog.Info("Validation for ChallengeInstance upon creation", "name", i.GetName())

	otherInstances := &prismctfv1.ChallengeInstanceList{}
	err := v.Client.List(ctx, otherInstances, client.MatchingFields{".spec.team": i.Spec.Team})
	if err != nil {
		challengeinstancelog.Error(err, "error listing team instances")
		return nil, err
	}
	if len(otherInstances.Items) >= *utils.MaxInstancesPerTeam {
		return nil, ErrLimitReached
	}
	for _, item := range otherInstances.Items {
		if item.Spec.Challenge == i.Spec.Challenge {
			return nil, ErrAlreadyExists
		}
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type ChallengeInstance.
func (v *ChallengeInstanceCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	challengeinstance, ok := newObj.(*prismctfv1.ChallengeInstance)
	if !ok {
		return nil, fmt.Errorf("expected a ChallengeInstance object for the newObj but got %T", newObj)
	}
	challengeinstancelog.Info("Validation for ChallengeInstance upon update", "name", challengeinstance.GetName())

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type ChallengeInstance.
func (v *ChallengeInstanceCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	challengeinstance, ok := obj.(*prismctfv1.ChallengeInstance)
	if !ok {
		return nil, fmt.Errorf("expected a ChallengeInstance object but got %T", obj)
	}
	challengeinstancelog.Info("Validation for ChallengeInstance upon deletion", "name", challengeinstance.GetName())

	return nil, nil
}
