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

package v1

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	prismctfv1 "github.com/pwnlentoni/prism-ctf/api/v1"
)

// nolint:unused
// log is for logging in this package.
var sharedchallengelog = logf.Log.WithName("sharedchallenge-resource")

// SetupSharedChallengeWebhookWithManager registers the webhook for SharedChallenge in the manager.
func SetupSharedChallengeWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &prismctfv1.SharedChallenge{}).
		WithValidator(&SharedChallengeCustomValidator{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: If you want to customise the 'path', use the flags '--defaulting-path' or '--validation-path'.
// +kubebuilder:webhook:path=/validate-prism-ctf-pwnlentoni-team-v1-sharedchallenge,mutating=false,failurePolicy=fail,sideEffects=None,groups=prism-ctf.pwnlentoni.team,resources=sharedchallenges,verbs=create;update,versions=v1,name=vsharedchallenge-v1.kb.io,admissionReviewVersions=v1

// SharedChallengeCustomValidator struct is responsible for validating the SharedChallenge resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type SharedChallengeCustomValidator struct {
}

func (v *SharedChallengeCustomValidator) validate(chall *prismctfv1.SharedChallenge) (warnings admission.Warnings, err error) {
	containers, err := validateContainers(chall.Spec.Containers, field.NewPath("spec", "containers"))
	if err != nil {
		return nil, err
	}

	err = validateExposures(containers, chall.Spec.Exposes, field.NewPath("spec", "exposes"))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type SharedChallenge.
func (v *SharedChallengeCustomValidator) ValidateCreate(_ context.Context, chall *prismctfv1.SharedChallenge) (admission.Warnings, error) {
	sharedchallengelog.Info("Validation for SharedChallenge upon creation", "name", chall.GetName())

	return v.validate(chall)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type SharedChallenge.
func (v *SharedChallengeCustomValidator) ValidateUpdate(_ context.Context, oldObj, chall *prismctfv1.SharedChallenge) (admission.Warnings, error) {
	sharedchallengelog.Info("Validation for SharedChallenge upon update", "name", chall.GetName())

	return v.validate(chall)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type SharedChallenge.
func (v *SharedChallengeCustomValidator) ValidateDelete(_ context.Context, obj *prismctfv1.SharedChallenge) (admission.Warnings, error) {
	sharedchallengelog.Info("Validation for SharedChallenge upon deletion", "name", obj.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
