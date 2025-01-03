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
	"errors"
	"fmt"
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
var isolatedchallengelog = logf.Log.WithName("isolatedchallenge-resource")

// SetupIsolatedChallengeWebhookWithManager registers the webhook for IsolatedChallenge in the manager.
func SetupIsolatedChallengeWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&prismctfv1.IsolatedChallenge{}).
		WithValidator(&IsolatedChallengeCustomValidator{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-prism-ctf-pwnlentoni-team-v1-isolatedchallenge,mutating=false,failurePolicy=fail,sideEffects=None,groups=prism-ctf.pwnlentoni.team,resources=isolatedchallenges,verbs=create;update,versions=v1,name=visolatedchallenge-v1.kb.io,admissionReviewVersions=v1

// IsolatedChallengeCustomValidator struct is responsible for validating the IsolatedChallenge resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type IsolatedChallengeCustomValidator struct {
	//TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &IsolatedChallengeCustomValidator{}

func (v *IsolatedChallengeCustomValidator) validate(chal *prismctfv1.IsolatedChallenge) (warnings admission.Warnings, err error) {
	warnings = make(admission.Warnings, 0)
	defer func() {
		if len(warnings) == 0 {
			warnings = nil
		}
	}()
	spec := chal.Spec
	if spec.Lifetime == nil {
		err = errors.New("missing lifetime")
		return
	}
	if spec.Lifetime.Duration < 5*time.Minute {
		warnings = append(warnings, "lifetime under suggested 5 minutes")
	}
	return
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type IsolatedChallenge.
func (v *IsolatedChallengeCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	chall, ok := obj.(*prismctfv1.IsolatedChallenge)
	if !ok {
		return nil, fmt.Errorf("expected a IsolatedChallenge object but got %T", obj)
	}
	isolatedchallengelog.Info("Validation for IsolatedChallenge upon creation", "name", chall.GetName())
	return v.validate(chall)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type IsolatedChallenge.
func (v *IsolatedChallengeCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	chall, ok := newObj.(*prismctfv1.IsolatedChallenge)
	if !ok {
		return nil, fmt.Errorf("expected a IsolatedChallenge object for the newObj but got %T", newObj)
	}
	isolatedchallengelog.Info("Validation for IsolatedChallenge upon update", "name", chall.GetName())
	return v.validate(chall)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type IsolatedChallenge.
func (v *IsolatedChallengeCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	isolatedchallenge, ok := obj.(*prismctfv1.IsolatedChallenge)
	if !ok {
		return nil, fmt.Errorf("expected a IsolatedChallenge object but got %T", obj)
	}
	isolatedchallengelog.Info("Validation for IsolatedChallenge upon deletion", "name", isolatedchallenge.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
