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
	"fmt"
	"github.com/pwnlentoni/prism-ctf/internal/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	prismctfv1 "github.com/pwnlentoni/prism-ctf/api/v1"
)

// nolint:unused
// log is for logging in this package.
var sharedchallengelog = logf.Log.WithName("sharedchallenge-resource")

// SetupSharedChallengeWebhookWithManager registers the webhook for SharedChallenge in the manager.
func SetupSharedChallengeWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&prismctfv1.SharedChallenge{}).
		WithValidator(&SharedChallengeCustomValidator{Client: mgr.GetClient()}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-prism-ctf-pwnlentoni-team-v1-sharedchallenge,mutating=false,failurePolicy=fail,sideEffects=None,groups=prism-ctf.pwnlentoni.team,resources=sharedchallenges,verbs=create;update,versions=v1,name=vsharedchallenge-v1.kb.io,admissionReviewVersions=v1

// SharedChallengeCustomValidator struct is responsible for validating the SharedChallenge resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type SharedChallengeCustomValidator struct {
	Client client.Client
}

var _ webhook.CustomValidator = &SharedChallengeCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type SharedChallenge.
func (v *SharedChallengeCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	chall, ok := obj.(*prismctfv1.SharedChallenge)
	if !ok {
		return nil, fmt.Errorf("expected a SharedChallenge object but got %T", obj)
	}
	sharedchallengelog.Info("Validation for SharedChallenge upon creation", "name", chall.GetName())

	doc, err := utils.RenderSharedTemplate(chall.Spec.Template, "challs.pwnlentoni.team")
	if err != nil {
		return nil, err
	}

	err = validateDoc(ctx, v.Client, sharedchallengelog, doc, utils.SharedChallengesNamespace)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type SharedChallenge.
func (v *SharedChallengeCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	chall, ok := newObj.(*prismctfv1.SharedChallenge)
	if !ok {
		return nil, fmt.Errorf("expected a SharedChallenge object for the newObj but got %T", newObj)
	}
	sharedchallengelog.Info("Validation for SharedChallenge upon update", "name", chall.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type SharedChallenge.
func (v *SharedChallengeCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	sharedchallenge, ok := obj.(*prismctfv1.SharedChallenge)
	if !ok {
		return nil, fmt.Errorf("expected a SharedChallenge object but got %T", obj)
	}
	sharedchallengelog.Info("Validation for SharedChallenge upon deletion", "name", sharedchallenge.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
