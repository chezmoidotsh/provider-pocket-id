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

// UserGroupBindingParameters are the configurable fields of a UserGroupBinding.
// +kubebuilder:validation:XValidation:rule="(has(self.userId) ? 1 : 0) + (self.userIdRef != null ? 1 : 0) + (self.userIdSelector != null ? 1 : 0) == 1",message="Exactly one of userId, userIdRef or userIdSelector must be specified."
// +kubebuilder:validation:XValidation:rule="(has(self.groupId) ? 1 : 0) + (self.groupIdRef != null ? 1 : 0) + (self.groupIdSelector != null ? 1 : 0) == 1",message="Exactly one of groupId, groupIdRef or groupIdSelector must be specified."
type UserGroupBindingParameters struct {
	// UserID is the ID of the user to add to a group.
	// The user must already exist in Pocket ID.
	// +optional
	UserID string `json:"userId"`

	// UserIDRef is a reference to a User resource to add to a group.
	// This creates a dependency on the referenced User resource.
	// +optional
	UserIDRef *xpv1.Reference `json:"userIdRef"`

	// UserIDSelector selects a User resource to add to a group.
	// +optional
	UserIDSelector *xpv1.Selector `json:"userIdSelector"`

	// GroupID is the ID of the group to add the user to.
	// The group must already exist in Pocket ID.
	// +optional
	GroupID string `json:"groupId"`

	// GroupIDRef is a reference to a Group resource to add the user to.
	// This creates a dependency on the referenced Group resource.
	// +optional
	GroupIDRef *xpv1.Reference `json:"groupIdRef"`

	// GroupIDSelector selects a Group resource to add the user to.
	// +optional
	GroupIDSelector *xpv1.Selector `json:"groupIdSelector"`
}

// UserGroupBindingObservation are the observable fields of a UserGroupBinding.
type UserGroupBindingObservation struct {
	// User contains the full user information.
	User UserObservation `json:"user"`

	// Group contains the full group information.
	Group GroupObservation `json:"group"`
}

// A UserGroupBindingSpec defines the desired state of a UserGroupBinding.
type UserGroupBindingSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       UserGroupBindingParameters `json:"forProvider"`
}

// A UserGroupBindingStatus represents the observed state of a UserGroupBinding.
type UserGroupBindingStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          UserGroupBindingObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A UserGroupBinding adds a user to a group in Pocket ID.
// Groups are used to organize users and control access to OIDC applications.
// Users can belong to multiple groups, and groups can be used in OIDC client
// configurations to restrict access based on group membership.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="USERNAME",type="string",JSONPath=".status.atProvider.user.username"
// +kubebuilder:printcolumn:name="GROUP-NAME",type="string",JSONPath=".status.atProvider.group.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,pocketid}
type UserGroupBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserGroupBindingSpec   `json:"spec"`
	Status UserGroupBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserGroupBindingList contains a list of UserGroupBinding
type UserGroupBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserGroupBinding `json:"items"`
}

// UserGroupBinding type metadata.
var (
	UserGroupBindingKind             = reflect.TypeOf(UserGroupBinding{}).Name()
	UserGroupBindingGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: UserGroupBindingKind}.String()
	UserGroupBindingKindAPIVersion   = UserGroupBindingKind + "." + SchemeGroupVersion.String()
	UserGroupBindingGroupVersionKind = SchemeGroupVersion.WithKind(UserGroupBindingKind)
)

func init() {
	SchemeBuilder.Register(&UserGroupBinding{}, &UserGroupBindingList{})
}
