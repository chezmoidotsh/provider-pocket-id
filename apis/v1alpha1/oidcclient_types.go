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

package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// OIDCClientParameters are the configurable fields of an OIDCClient.
type OIDCClientParameters struct {
	// Name is the display name of the OIDC client application.
	// This is shown to users during the authentication flow.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// ID is the client identifier for OIDC. If not specified, a UUID will be generated.
	// +optional
	ID string `json:"id"`

	// CallbackURLs are the allowed redirect URIs after successful authentication.
	// These must be exact matches for security purposes.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:items:Format=uri
	CallbackURLs []string `json:"callbackURLs"`

	// LogoutCallbackURLs are the allowed redirect URIs after logout.
	// +optional
	LogoutCallbackURLs []string `json:"logoutCallbackURLs"`

	// LaunchURL is the application's main URL, used for display purposes.
	// +optional
	LaunchURL string `json:"launchURL"`

	// IsPublic indicates whether this is a public client (cannot keep secrets secure).
	// Public clients don't use client secrets and must use PKCE.
	// +optional
	IsPublic bool `json:"isPublic"`

	// PkceEnabled indicates whether Proof Key for Code Exchange is required.
	// This should be enabled for enhanced security, especially for public clients.
	// +optional
	PkceEnabled bool `json:"pkceEnabled"`

	// RequiresReauthentication forces users to re-authenticate even if they have an active session.
	// +optional
	RequiresReauthentication bool `json:"requiresReauthentication"`

	// LogoURL is the URL to an image file that will be used as the client's logo.
	// The provider will download this image and upload it to Pocket ID.
	// Supported formats: PNG, JPEG, GIF, SVG. Maximum size: 2MB.
	// +optional
	// +kubebuilder:validation:Format=uri
	LogoURL string `json:"logoUrl"`

	// Credentials configure federated client authentication methods.
	// +optional
	Credentials OIDCClientCredentials `json:"credentials"`
}

// OIDCClientCredentials are the configurable fields of an OIDCClient's credentials.
type OIDCClientCredentials struct {
	// FederatedIdentities configure JWT-based client authentication.
	// This allows clients to authenticate using JWTs from trusted issuers.
	// +optional
	FederatedIdentities []OIDCClientCredentialsFederatedIdentity `json:"federatedIdentities"`
}

// OIDCClientCredentialsFederatedIdentity are the configurable fields of an OIDCClient's federated identity credentials.
type OIDCClientCredentialsFederatedIdentity struct {
	// Issuer must match the 'iss' claim in JWT tokens from the external IdP.
	// This identifies the trusted token issuer.
	Issuer string `json:"issuer"`

	// Subject must match the 'sub' claim in JWT tokens.
	// If empty, defaults to the OIDC client UUID.
	// +optional
	Subject string `json:"subject"`

	// Audience must match the 'aud' claim in JWT tokens.
	// This verifies the intended recipient of the token.
	// +optional
	Audience string `json:"audience"`

	// JWKS is the JSON Web Key Set URL for verifying JWT signatures.
	// If empty, defaults to <issuer>/.well-known/jwks.json
	// +optional
	JWKS string `json:"jwks"`
}

// OIDCClientObservation are the observable fields of an OIDCClient.
type OIDCClientObservation struct {
	// ID is the unique identifier of the OIDC client in Pocket ID.
	ID string `json:"id"`

	// Name is the display name of the OIDC client application.
	Name string `json:"name"`

	// CallbackURLs are the configured redirect URIs.
	CallbackURLs []string `json:"callbackURLs,omitempty"`

	// LogoutCallbackURLs are the configured logout redirect URIs.
	LogoutCallbackURLs []string `json:"logoutCallbackURLs,omitempty"`

	// LaunchURL is the application's main URL.
	LaunchURL string `json:"launchURL,omitempty"`

	// IsPublic indicates if this is a public client.
	IsPublic bool `json:"isPublic,omitempty"`

	// PkceEnabled indicates if PKCE is required.
	PkceEnabled bool `json:"pkceEnabled,omitempty"`

	// RequiresReauthentication indicates if re-authentication is required.
	RequiresReauthentication bool `json:"requiresReauthentication,omitempty"`

	// LogoURL is the configured logo URL for this client.
	LogoURL string `json:"logoUrl,omitempty"`

	// HasLogo indicates whether a logo has been uploaded for this client.
	HasLogo bool `json:"hasLogo,omitempty"`

	// Credentials contain the federated authentication configuration.
	Credentials OIDCClientCredentials `json:"credentials,omitempty"`
}

// An OIDCClientSpec defines the desired state of an OIDCClient.
type OIDCClientSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       OIDCClientParameters `json:"forProvider"`
}

// An OIDCClientStatus represents the observed state of an OIDCClient.
type OIDCClientStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          OIDCClientObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An OIDCClient represents an OIDC client application in Pocket ID.
// OIDC clients are applications that can request authentication from Pocket ID
// and receive user identity information through OpenID Connect protocols.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="CLIENT-NAME",type="string",JSONPath=".status.atProvider.name"
// +kubebuilder:printcolumn:name="PUBLIC",type="boolean",JSONPath=".status.atProvider.isPublic"
// +kubebuilder:printcolumn:name="PKCE",type="boolean",JSONPath=".status.atProvider.pkceEnabled"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,pocketid}
type OIDCClient struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OIDCClientSpec   `json:"spec"`
	Status OIDCClientStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OIDCClientList contains a list of OIDCClient
type OIDCClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OIDCClient `json:"items"`
}

// OIDCClient type metadata.
var (
	OIDCClientKind             = reflect.TypeOf(OIDCClient{}).Name()
	OIDCClientGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: OIDCClientKind}.String()
	OIDCClientKindAPIVersion   = OIDCClientKind + "." + SchemeGroupVersion.String()
	OIDCClientGroupVersionKind = SchemeGroupVersion.WithKind(OIDCClientKind)
)

func init() {
	SchemeBuilder.Register(&OIDCClient{}, &OIDCClientList{})
}
