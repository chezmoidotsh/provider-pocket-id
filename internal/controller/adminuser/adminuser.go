/*
Copyright 2025 The Crossplane Authors.

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

package adminuser

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/feature"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/statemetrics"

	apisv1alpha1 "github.com/crossplane/provider-pocketid/apis/v1alpha1"
	"github.com/crossplane/provider-pocketid/internal/clients/pocketid"
	"github.com/crossplane/provider-pocketid/internal/features"
)

const (
	errNotAdminUser = "managed resource is not a AdminUser custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

	errNewClient = "cannot create new Service"
)

// newPocketIDService creates a new Pocket ID service
var (
	newPocketIDService = func(endpoint string, creds []byte) (interface{}, error) {
		return pocketid.NewClientFromCredentials(endpoint, string(creds))
	}
)

// Setup adds a controller that reconciles AdminUser managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(apisv1alpha1.AdminUserGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	opts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: newPocketIDService}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
		managed.WithManagementPolicies(),
	}

	if o.Features.Enabled(feature.EnableAlphaChangeLogs) {
		opts = append(opts, managed.WithChangeLogger(o.ChangeLogOptions.ChangeLogger))
	}

	if o.MetricOptions != nil {
		opts = append(opts, managed.WithMetricRecorder(o.MetricOptions.MRMetrics))
	}

	if o.MetricOptions != nil && o.MetricOptions.MRStateMetrics != nil {
		stateMetricsRecorder := statemetrics.NewMRStateRecorder(
			mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &apisv1alpha1.AdminUserList{}, o.MetricOptions.PollStateMetricInterval,
		)
		if err := mgr.Add(stateMetricsRecorder); err != nil {
			return errors.Wrap(err, "cannot register MR state metrics recorder for kind apisv1alpha1.AdminUserList")
		}
	}

	r := managed.NewReconciler(mgr, resource.ManagedKind(apisv1alpha1.AdminUserGroupVersionKind), opts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&apisv1alpha1.AdminUser{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(endpoint string, creds []byte) (interface{}, error)
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*apisv1alpha1.AdminUser)
	if !ok {
		return nil, errors.New(errNotAdminUser)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	cd := pc.Spec.Credentials
	data, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	svc, err := c.newServiceFn(pc.Spec.Endpoint, data)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{service: svc.(*pocketid.Client)}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	service *pocketid.Client
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*apisv1alpha1.AdminUser)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotAdminUser)
	}

	// Use external-name annotation if present, otherwise use username
	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		externalName = cr.Spec.ForProvider.Username
	}

	user, err := c.service.GetUserByExternalName(ctx, externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get admin user")
	}

	if user == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	// Ensure this user is actually an admin
	if !user.IsAdmin {
		return managed.ExternalObservation{}, errors.New("user exists but is not an admin user")
	}

	// Update status with observed values
	cr.Status.AtProvider = apisv1alpha1.AdminUserObservation{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Locale:       user.Locale,
		Disabled:     user.Disabled,
		IsAdmin:      user.IsAdmin,
		UserGroups:   user.UserGroups,
		CustomClaims: user.CustomClaims,
	}

	// Set external name to username if not already set
	if meta.GetExternalName(cr) == "" {
		meta.SetExternalName(cr, user.Username)
	}

	// Check if resource is up to date
	upToDate := isAdminUserUpToDate(cr.Spec.ForProvider, *user)

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*apisv1alpha1.AdminUser)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotAdminUser)
	}

	req := pocketid.CreateUserRequest{
		Username:     cr.Spec.ForProvider.Username,
		Email:        cr.Spec.ForProvider.Email,
		FirstName:    cr.Spec.ForProvider.FirstName,
		LastName:     cr.Spec.ForProvider.LastName,
		Locale:       cr.Spec.ForProvider.Locale,
		Disabled:     cr.Spec.ForProvider.Disabled,
		IsAdmin:      true, // AdminUser resources create admin users
		CustomClaims: cr.Spec.ForProvider.CustomClaims,
	}

	user, err := c.service.CreateUser(ctx, req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create admin user")
	}

	// Set external name to username
	meta.SetExternalName(cr, user.Username)

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*apisv1alpha1.AdminUser)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotAdminUser)
	}

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalUpdate{}, errors.New("admin user ID not found in status")
	}

	req := pocketid.UpdateUserRequest{
		Username:     cr.Spec.ForProvider.Username,
		Email:        cr.Spec.ForProvider.Email,
		FirstName:    cr.Spec.ForProvider.FirstName,
		LastName:     cr.Spec.ForProvider.LastName,
		Locale:       cr.Spec.ForProvider.Locale,
		Disabled:     cr.Spec.ForProvider.Disabled,
		CustomClaims: cr.Spec.ForProvider.CustomClaims,
	}

	_, err := c.service.UpdateUser(ctx, cr.Status.AtProvider.ID, req)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update admin user")
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*apisv1alpha1.AdminUser)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotAdminUser)
	}

	if cr.Status.AtProvider.ID != "" {
		err := c.service.DeleteUser(ctx, cr.Status.AtProvider.ID)
		if err != nil {
			return managed.ExternalDelete{}, errors.Wrap(err, "failed to delete admin user")
		}
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	return nil
}

// isAdminUserUpToDate compares the desired spec with the actual admin user state
//
//nolint:gocyclo
func isAdminUserUpToDate(spec apisv1alpha1.AdminUserParameters, user pocketid.User) bool {
	if spec.Username != user.Username {
		return false
	}
	if spec.Email != user.Email {
		return false
	}
	if spec.FirstName != user.FirstName {
		return false
	}
	if spec.LastName != user.LastName {
		return false
	}
	if spec.Locale != user.Locale {
		return false
	}
	if spec.Disabled != user.Disabled {
		return false
	}

	// Compare custom claims
	if len(spec.CustomClaims) != len(user.CustomClaims) {
		return false
	}
	for k, v := range spec.CustomClaims {
		if userVal, exists := user.CustomClaims[k]; !exists || userVal != v {
			return false
		}
	}

	return true
}
