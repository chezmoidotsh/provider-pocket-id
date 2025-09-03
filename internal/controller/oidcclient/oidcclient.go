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

package oidcclient

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
	errNotOIDCClient = "managed resource is not an OIDCClient custom resource"
	errTrackPCUsage  = "cannot track ProviderConfig usage"
	errGetPC         = "cannot get ProviderConfig"
	errGetCreds      = "cannot get credentials"

	errNewClient = "cannot create new Service"
)

// newPocketIDService creates a new Pocket ID service
var (
	newPocketIDService = func(endpoint string, creds []byte) (interface{}, error) {
		return pocketid.NewClientFromCredentials(endpoint, string(creds))
	}
)

// Setup adds a controller that reconciles Client managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(apisv1alpha1.OIDCClientGroupKind)

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
			mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &apisv1alpha1.OIDCClientList{}, o.MetricOptions.PollStateMetricInterval,
		)
		if err := mgr.Add(stateMetricsRecorder); err != nil {
			return errors.Wrap(err, "cannot register MR state metrics recorder for kind apisv1alpha1.OIDCClientList")
		}
	}

	r := managed.NewReconciler(mgr, resource.ManagedKind(apisv1alpha1.OIDCClientGroupVersionKind), opts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&apisv1alpha1.OIDCClient{}).
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
	cr, ok := mg.(*apisv1alpha1.OIDCClient)
	if !ok {
		return nil, errors.New(errNotOIDCClient)
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
	cr, ok := mg.(*apisv1alpha1.OIDCClient)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotOIDCClient)
	}

	// Use external-name annotation if present, otherwise use name
	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		externalName = cr.Spec.ForProvider.Name
	}

	client, err := c.service.GetOIDCClientByExternalName(ctx, externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get OIDC client")
	}

	if client == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	// Update status with observed values
	cr.Status.AtProvider = apisv1alpha1.OIDCClientObservation{
		ID:                 client.ID,
		Name:               client.ClientName,
		CallbackURLs:       client.RedirectURIs,
		LogoutCallbackURLs: client.PostLogoutURIs,
		LaunchURL:          client.LaunchURL,
		IsPublic:           client.IsPublic,
		PkceEnabled:        client.RequirePKCE,
		HasLogo:            client.HasLogo,
	}

	// Set external name to clientName if not already set
	if meta.GetExternalName(cr) == "" {
		meta.SetExternalName(cr, client.ClientName)
	}

	// Check if resource is up to date
	upToDate := isOIDCClientUpToDate(cr.Spec.ForProvider, *client)

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*apisv1alpha1.OIDCClient)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotOIDCClient)
	}

	req := pocketid.CreateOIDCClientRequest{
		ClientName:     cr.Spec.ForProvider.Name,
		RedirectURIs:   cr.Spec.ForProvider.CallbackURLs,
		PostLogoutURIs: cr.Spec.ForProvider.LogoutCallbackURLs,
		LaunchURL:      cr.Spec.ForProvider.LaunchURL,
		IsPublic:       cr.Spec.ForProvider.IsPublic,
		RequirePKCE:    cr.Spec.ForProvider.PkceEnabled,
	}

	client, err := c.service.CreateOIDCClient(ctx, req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create OIDC client")
	}

	// Set external name to clientName
	meta.SetExternalName(cr, client.ClientName)

	// Handle logo upload if specified
	if cr.Spec.ForProvider.LogoURL != "" {
		//nolint:staticcheck
		if err := c.service.UploadOIDCClientLogo(ctx, client.ID, cr.Spec.ForProvider.LogoURL); err != nil {
			// Log the error but don't fail the creation
			// The logo can be uploaded later during update
		}
	}

	// Return client secret as connection detail if not public
	connectionDetails := managed.ConnectionDetails{}
	if !client.IsPublic && client.ClientSecret != "" {
		connectionDetails["clientSecret"] = []byte(client.ClientSecret)
	}

	return managed.ExternalCreation{
		ConnectionDetails: connectionDetails,
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*apisv1alpha1.OIDCClient)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotOIDCClient)
	}

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalUpdate{}, errors.New("OIDC client ID not found in status")
	}

	req := pocketid.UpdateOIDCClientRequest{
		ClientName:     cr.Spec.ForProvider.Name,
		RedirectURIs:   cr.Spec.ForProvider.CallbackURLs,
		PostLogoutURIs: cr.Spec.ForProvider.LogoutCallbackURLs,
		LaunchURL:      cr.Spec.ForProvider.LaunchURL,
		IsPublic:       cr.Spec.ForProvider.IsPublic,
		RequirePKCE:    cr.Spec.ForProvider.PkceEnabled,
	}

	_, err := c.service.UpdateOIDCClient(ctx, cr.Status.AtProvider.ID, req)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update OIDC client")
	}

	// Handle logo upload if specified and different from current state
	if cr.Spec.ForProvider.LogoURL != "" {
		// Always try to upload logo on update - API will handle if it's the same
		//nolint:staticcheck
		if err := c.service.UploadOIDCClientLogo(ctx, cr.Status.AtProvider.ID, cr.Spec.ForProvider.LogoURL); err != nil {
			// Log the error but don't fail the update
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*apisv1alpha1.OIDCClient)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotOIDCClient)
	}

	if cr.Status.AtProvider.ID != "" {
		err := c.service.DeleteOIDCClient(ctx, cr.Status.AtProvider.ID)
		if err != nil {
			return managed.ExternalDelete{}, errors.Wrap(err, "failed to delete OIDC client")
		}
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	return nil
}

// isOIDCClientUpToDate compares the desired spec with the actual OIDC client state
func isOIDCClientUpToDate(spec apisv1alpha1.OIDCClientParameters, client pocketid.OIDCClient) bool {
	if spec.Name != client.ClientName {
		return false
	}
	if spec.LaunchURL != client.LaunchURL {
		return false
	}
	if spec.IsPublic != client.IsPublic {
		return false
	}
	if spec.PkceEnabled != client.RequirePKCE {
		return false
	}

	// Compare string slices
	if !equalStringSlices(spec.CallbackURLs, client.RedirectURIs) {
		return false
	}
	if !equalStringSlices(spec.LogoutCallbackURLs, client.PostLogoutURIs) {
		return false
	}

	// Logo is handled separately and doesn't affect up-to-date status
	// since logos are uploaded asynchronously

	return true
}

// equalStringSlices compares two string slices for equality
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps to count occurrences
	countA := make(map[string]int)
	countB := make(map[string]int)

	for _, item := range a {
		countA[item]++
	}
	for _, item := range b {
		countB[item]++
	}

	// Compare maps
	for k, v := range countA {
		if countB[k] != v {
			return false
		}
	}

	return true
}
