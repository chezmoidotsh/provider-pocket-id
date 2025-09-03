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

// UserParameters are the configurable fields of a User.
type UserParameters struct {
	// Username is the unique username for the user account.
	// This is used for identification and must be unique within Pocket ID.
	Username string `json:"username"`

	// Email is the user's email address.
	// This is required for authentication and communication purposes.
	Email string `json:"email"`

	// FirstName is the user's given name.
	FirstName string `json:"firstName"`

	// LastName is the user's family name.
	// +optional
	LastName string `json:"lastName"`

	// Locale specifies the user's preferred language and region (e.g., "en-US", "fr-FR").
	// This affects the language used in Pocket ID interfaces and communications.
	// +optional
	Locale string `json:"locale"`

	// Disabled indicates whether the user account is disabled.
	// Disabled users cannot authenticate or access any services.
	// +optional
	Disabled bool `json:"disabled"`

	// CustomClaims are additional key-value pairs that will be included in JWT tokens.
	// These can be used to pass custom information to OIDC clients.
	// +optional
	CustomClaims map[string]string `json:"customClaims"`
}

// UserObservation are the observable fields of a User.
type UserObservation struct {
	// ID is the unique identifier of the user in Pocket ID.
	ID string `json:"id"`

	// Username is the user's username.
	Username string `json:"username"`

	// Email is the user's email address.
	Email string `json:"email"`

	// FirstName is the user's given name.
	FirstName string `json:"firstName"`

	// LastName is the user's family name.
	LastName string `json:"lastName,omitempty"`

	// Locale is the user's preferred language and region.
	Locale string `json:"locale,omitempty"`

	// Disabled indicates whether the user account is disabled.
	Disabled bool `json:"disabled,omitempty"`

	// IsAdmin indicates whether this user has administrative privileges.
	// This is read-only and managed separately through AdminUser resources.
	IsAdmin bool `json:"isAdmin,omitempty"`

	// UserGroups lists the names of groups this user belongs to.
	// This is managed through UserGroupBinding resources.
	UserGroups []string `json:"userGroups,omitempty"`

	// CustomClaims are the custom key-value pairs included in JWT tokens.
	CustomClaims map[string]string `json:"customClaims,omitempty"`
}

// A UserSpec defines the desired state of a User.
type UserSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       UserParameters `json:"forProvider"`
}

// A UserStatus represents the observed state of a User.
type UserStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          UserObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A User represents a regular user account in Pocket ID.
// Users can authenticate using passkeys and access applications through OIDC.
// Admin privileges are managed separately through AdminUser resources.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="USERNAME",type="string",JSONPath=".status.atProvider.username"
// +kubebuilder:printcolumn:name="EMAIL",type="string",JSONPath=".status.atProvider.email"
// +kubebuilder:printcolumn:name="DISABLED",type="boolean",JSONPath=".status.atProvider.disabled"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,pocketid}
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec"`
	Status UserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserList contains a list of User
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

// User type metadata.
var (
	UserKind             = reflect.TypeOf(User{}).Name()
	UserGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: UserKind}.String()
	UserKindAPIVersion   = UserKind + "." + SchemeGroupVersion.String()
	UserGroupVersionKind = SchemeGroupVersion.WithKind(UserKind)
)

func init() {
	SchemeBuilder.Register(&User{}, &UserList{})
}
