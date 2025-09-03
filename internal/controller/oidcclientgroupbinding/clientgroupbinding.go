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

package clientgroupbinding

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
	errNotClientGroupBinding = "managed resource is not a OIDCClientGroupBinding custom resource"
	errTrackPCUsage          = "cannot track ProviderConfig usage"
	errGetPC                 = "cannot get ProviderConfig"
	errGetCreds              = "cannot get credentials"
	errNewClient             = "cannot create new Service"
	errResolveClientID       = "cannot resolve client ID"
	errResolveGroupID        = "cannot resolve group ID"
)

// newPocketIDService creates a new Pocket ID service
var (
	newPocketIDService = func(endpoint string, creds []byte) (interface{}, error) {
		return pocketid.NewClientFromCredentials(endpoint, string(creds))
	}
)

// Setup adds a controller that reconciles OIDCClientGroupBinding managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(apisv1alpha1.OIDCClientGroupBindingGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	opts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: newPocketIDService,
		}),
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
			mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &apisv1alpha1.OIDCClientGroupBindingList{}, o.MetricOptions.PollStateMetricInterval,
		)
		if err := mgr.Add(stateMetricsRecorder); err != nil {
			return errors.Wrap(err, "cannot register MR state metrics recorder for kind apisv1alpha1.OIDCClientGroupBindingList")
		}
	}

	r := managed.NewReconciler(mgr, resource.ManagedKind(apisv1alpha1.OIDCClientGroupBindingGroupVersionKind), opts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&apisv1alpha1.OIDCClientGroupBinding{}).
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
	cr, ok := mg.(*apisv1alpha1.OIDCClientGroupBinding)
	if !ok {
		return nil, errors.New(errNotClientGroupBinding)
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

	return &external{
		service: svc.(*pocketid.Client),
		kube:    c.kube,
	}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	service *pocketid.Client
	kube    client.Client
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*apisv1alpha1.OIDCClientGroupBinding)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotClientGroupBinding)
	}

	// Resolve client ID
	clientID, err := c.resolveClientID(ctx, cr)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errResolveClientID)
	}

	// Resolve group ID
	groupID, err := c.resolveGroupID(ctx, cr)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errResolveGroupID)
	}

	// Check if binding exists
	exists, err := c.service.IsClientInGroup(ctx, clientID, groupID)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to check client group binding")
	}

	if !exists {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	// Get client and group details for status
	client, err := c.service.GetOIDCClient(ctx, clientID)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get OIDC client")
	}

	group, err := c.service.GetGroup(ctx, groupID)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get group")
	}

	// Update status with observed values
	cr.Status.AtProvider = apisv1alpha1.OIDCClientGroupBindingObservation{
		Client: apisv1alpha1.OIDCClientObservation{
			ID:                 client.ID,
			Name:               client.ClientName,
			CallbackURLs:       client.RedirectURIs,
			LogoutCallbackURLs: client.PostLogoutURIs,
			LaunchURL:          client.LaunchURL,
			IsPublic:           client.IsPublic,
			PkceEnabled:        client.RequirePKCE,
			HasLogo:            client.HasLogo,
		},
		Group: apisv1alpha1.GroupObservation{
			ID:           group.ID,
			Name:         group.GroupName,
			FriendlyName: group.FriendlyName,
			CustomClaims: group.CustomClaims,
		},
	}

	// Set external name combining client and group IDs
	if meta.GetExternalName(cr) == "" {
		meta.SetExternalName(cr, clientID+":"+groupID)
	}

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true, // Bindings don't have updatable fields
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*apisv1alpha1.OIDCClientGroupBinding)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotClientGroupBinding)
	}

	// Resolve client ID
	clientID, err := c.resolveClientID(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errResolveClientID)
	}

	// Resolve group ID
	groupID, err := c.resolveGroupID(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errResolveGroupID)
	}

	// Add client to group
	err = c.service.AddClientToGroup(ctx, clientID, groupID)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create client group binding")
	}

	// Set external name combining client and group IDs
	meta.SetExternalName(cr, clientID+":"+groupID)

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// Bindings don't have updatable fields, so this is essentially a no-op
	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*apisv1alpha1.OIDCClientGroupBinding)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotClientGroupBinding)
	}

	// Resolve client ID
	clientID, err := c.resolveClientID(ctx, cr)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, errResolveClientID)
	}

	// Resolve group ID
	groupID, err := c.resolveGroupID(ctx, cr)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, errResolveGroupID)
	}

	// Remove client from group
	err = c.service.RemoveClientFromGroup(ctx, clientID, groupID)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, "failed to delete client group binding")
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	return nil
}

// resolveClientID resolves the client ID from the binding spec
func (c *external) resolveClientID(ctx context.Context, cr *apisv1alpha1.OIDCClientGroupBinding) (string, error) {
	if cr.Spec.ForProvider.ClientID != "" {
		return cr.Spec.ForProvider.ClientID, nil
	}

	if cr.Spec.ForProvider.ClientIDRef != nil {
		oidcClient := &apisv1alpha1.OIDCClient{}
		if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.Spec.ForProvider.ClientIDRef.Name}, oidcClient); err != nil {
			return "", errors.Wrap(err, "failed to get referenced OIDC client")
		}
		if oidcClient.Status.AtProvider.ID == "" {
			return "", errors.New("referenced OIDC client ID is not available")
		}
		return oidcClient.Status.AtProvider.ID, nil
	}

	// TODO: Implement selector logic if needed
	return "", errors.New("client ID, clientIdRef, or clientIdSelector must be specified")
}

// resolveGroupID resolves the group ID from the binding spec
func (c *external) resolveGroupID(ctx context.Context, cr *apisv1alpha1.OIDCClientGroupBinding) (string, error) {
	if cr.Spec.ForProvider.GroupID != "" {
		return cr.Spec.ForProvider.GroupID, nil
	}

	if cr.Spec.ForProvider.GroupIDRef != nil {
		group := &apisv1alpha1.Group{}
		if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.Spec.ForProvider.GroupIDRef.Name}, group); err != nil {
			return "", errors.Wrap(err, "failed to get referenced group")
		}
		if group.Status.AtProvider.ID == "" {
			return "", errors.New("referenced group ID is not available")
		}
		return group.Status.AtProvider.ID, nil
	}

	// TODO: Implement selector logic if needed
	return "", errors.New("group ID, groupIdRef, or groupIdSelector must be specified")
}
