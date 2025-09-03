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

// GroupParameters are the configurable fields of a Group.
type GroupParameters struct {
	// Name is the unique identifier for the group.
	// This is used internally and must be unique within Pocket ID.
	Name string `json:"name"`

	// FriendlyName is the display name for the group.
	// This is shown to users and administrators in the Pocket ID interface.
	FriendlyName string `json:"friendlyName"`

	// CustomClaims are additional key-value pairs that will be included in JWT tokens
	// for users who belong to this group. These can be used to pass custom
	// information to OIDC clients based on group membership.
	// +optional
	CustomClaims map[string]string `json:"customClaims"`
}

// GroupObservation are the observable fields of a Group.
type GroupObservation struct {
	// ID is the unique identifier of the group in Pocket ID.
	ID string `json:"id"`

	// Name is the group's unique name.
	Name string `json:"name"`

	// FriendlyName is the group's display name.
	FriendlyName string `json:"friendlyName"`

	// CreatedAt is the timestamp when the group was created.
	CreatedAt string `json:"createdAt,omitempty"`

	// CustomClaims are the custom key-value pairs included in JWT tokens for group members.
	CustomClaims map[string]string `json:"customClaims,omitempty"`
}

// A GroupSpec defines the desired state of a Group.
type GroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       GroupParameters `json:"forProvider"`
}

// A GroupStatus represents the observed state of a Group.
type GroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          GroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Group represents a collection of users in Pocket ID.
// Groups are used to organize users and control access to OIDC applications.
// Users can be added to groups via UserGroupBinding resources, and groups
// can be associated with OIDC clients via OIDCClientGroupBinding resources
// to restrict application access based on group membership.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="GROUP-NAME",type="string",JSONPath=".status.atProvider.name"
// +kubebuilder:printcolumn:name="FRIENDLY-NAME",type="string",JSONPath=".status.atProvider.friendlyName"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,pocketid}
type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupSpec   `json:"spec"`
	Status GroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GroupList contains a list of Group
type GroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Group `json:"items"`
}

// Group type metadata.
var (
	GroupKind             = reflect.TypeOf(Group{}).Name()
	GroupGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: GroupKind}.String()
	GroupKindAPIVersion   = GroupKind + "." + SchemeGroupVersion.String()
	GroupGroupVersionKind = SchemeGroupVersion.WithKind(GroupKind)
)

func init() {
	SchemeBuilder.Register(&Group{}, &GroupList{})
}
