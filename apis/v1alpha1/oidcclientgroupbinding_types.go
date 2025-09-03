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

// OIDCClientGroupBindingParameters are the configurable fields of an OIDCClientGroupBinding.
type OIDCClientGroupBindingParameters struct {
	// ClientID is the ID of the OIDC client to bind to a group.
	// The client must already exist in Pocket ID.
	// +optional
	ClientID string `json:"clientId"`

	// ClientIDRef is a reference to an OIDCClient resource to bind to a group.
	// This creates a dependency on the referenced OIDCClient resource.
	// +optional
	ClientIDRef *xpv1.Reference `json:"clientIdRef"`

	// ClientIDSelector selects an OIDCClient resource to bind to a group.
	// +optional
	ClientIDSelector *xpv1.Selector `json:"clientIdSelector"`

	// GroupID is the ID of the group to bind the OIDC client to.
	// The group must already exist in Pocket ID.
	// +optional
	GroupID string `json:"groupId"`

	// GroupIDRef is a reference to a Group resource to bind the client to.
	// This creates a dependency on the referenced Group resource.
	// +optional
	GroupIDRef *xpv1.Reference `json:"groupIdRef"`

	// GroupIDSelector selects a Group resource to bind the client to.
	// +optional
	GroupIDSelector *xpv1.Selector `json:"groupIdSelector"`
}

// OIDCClientGroupBindingObservation are the observable fields of an OIDCClientGroupBinding.
type OIDCClientGroupBindingObservation struct {
	// Client contains the full OIDC client information.
	Client OIDCClientObservation `json:"client"`

	// Group contains the full group information.
	Group GroupObservation `json:"group"`
}

// An OIDCClientGroupBindingSpec defines the desired state of an OIDCClientGroupBinding.
type OIDCClientGroupBindingSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       OIDCClientGroupBindingParameters `json:"forProvider"`
}

// An OIDCClientGroupBindingStatus represents the observed state of an OIDCClientGroupBinding.
type OIDCClientGroupBindingStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          OIDCClientGroupBindingObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An OIDCClientGroupBinding associates an OIDC client with a group in Pocket ID.
// This allows you to restrict which users (based on their group membership)
// can access specific OIDC applications. Only users who belong to the bound
// group will be able to authenticate to the OIDC client.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="CLIENT-NAME",type="string",JSONPath=".status.atProvider.client.name"
// +kubebuilder:printcolumn:name="GROUP-NAME",type="string",JSONPath=".status.atProvider.group.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,pocketid}
type OIDCClientGroupBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OIDCClientGroupBindingSpec   `json:"spec"`
	Status OIDCClientGroupBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OIDCClientGroupBindingList contains a list of OIDCClientGroupBinding
type OIDCClientGroupBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OIDCClientGroupBinding `json:"items"`
}

// OIDCClientGroupBinding type metadata.
var (
	OIDCClientGroupBindingKind             = reflect.TypeOf(OIDCClientGroupBinding{}).Name()
	OIDCClientGroupBindingGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: OIDCClientGroupBindingKind}.String()
	OIDCClientGroupBindingKindAPIVersion   = OIDCClientGroupBindingKind + "." + SchemeGroupVersion.String()
	OIDCClientGroupBindingGroupVersionKind = SchemeGroupVersion.WithKind(OIDCClientGroupBindingKind)
)

func init() {
	SchemeBuilder.Register(&OIDCClientGroupBinding{}, &OIDCClientGroupBindingList{})
}
