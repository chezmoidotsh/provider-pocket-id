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

// AdminUserParameters are the configurable fields of an AdminUser.
// These are identical to UserParameters as AdminUser creates a user with admin privileges.
type AdminUserParameters struct {
	// Username is the unique username for the admin user account.
	// This is used for identification and must be unique within Pocket ID.
	// +kubebuilder:validation:Required
	Username string `json:"username"`

	// Email is the admin user's email address.
	// This is required for authentication and communication purposes.
	// +kubebuilder:validation:Required
	Email string `json:"email"`

	// FirstName is the admin user's given name.
	// +kubebuilder:validation:Required
	FirstName string `json:"firstName"`

	// LastName is the admin user's family name.
	// +optional
	LastName string `json:"lastName"`

	// Locale specifies the admin user's preferred language and region (e.g., "en-US", "fr-FR").
	// This affects the language used in Pocket ID interfaces and communications.
	// +optional
	Locale string `json:"locale"`

	// Disabled indicates whether the admin user account is disabled.
	// Disabled admin users cannot authenticate or access any services.
	// +optional
	Disabled bool `json:"disabled"`

	// CustomClaims are additional key-value pairs that will be included in JWT tokens.
	// These can be used to pass custom information to OIDC clients.
	// +optional
	CustomClaims map[string]string `json:"customClaims"`
}

// AdminUserObservation are the observable fields of an AdminUser.
// These are identical to UserObservation as AdminUser is a user with admin privileges.
type AdminUserObservation struct {
	// ID is the unique identifier of the admin user in Pocket ID.
	ID string `json:"id"`

	// Username is the admin user's username.
	Username string `json:"username"`

	// Email is the admin user's email address.
	Email string `json:"email"`

	// FirstName is the admin user's given name.
	FirstName string `json:"firstName"`

	// LastName is the admin user's family name.
	LastName string `json:"lastName,omitempty"`

	// Locale is the admin user's preferred language and region.
	Locale string `json:"locale,omitempty"`

	// Disabled indicates whether the admin user account is disabled.
	Disabled bool `json:"disabled,omitempty"`

	// IsAdmin will always be true for AdminUser resources.
	IsAdmin bool `json:"isAdmin,omitempty"`

	// UserGroups lists the names of groups this admin user belongs to.
	// This is managed through UserGroupBinding resources.
	UserGroups []string `json:"userGroups,omitempty"`

	// CustomClaims are the custom key-value pairs included in JWT tokens.
	CustomClaims map[string]string `json:"customClaims,omitempty"`
}

// An AdminUserSpec defines the desired state of an AdminUser.
type AdminUserSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       AdminUserParameters `json:"forProvider"`
}

// An AdminUserStatus represents the observed state of an AdminUser.
type AdminUserStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          AdminUserObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An AdminUser represents an administrator user account in Pocket ID.
// AdminUsers are created with administrative privileges (isAdmin: true) and can
// access the Pocket ID administrative interface to manage other users, groups,
// and OIDC clients. This is functionally identical to User except that the
// user is created with admin privileges from the start.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="USERNAME",type="string",JSONPath=".status.atProvider.username"
// +kubebuilder:printcolumn:name="EMAIL",type="string",JSONPath=".status.atProvider.email"
// +kubebuilder:printcolumn:name="DISABLED",type="boolean",JSONPath=".status.atProvider.disabled"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,pocketid}
type AdminUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AdminUserSpec   `json:"spec"`
	Status AdminUserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AdminUserList contains a list of AdminUser
type AdminUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AdminUser `json:"items"`
}

// AdminUser type metadata.
var (
	AdminUserKind             = reflect.TypeOf(AdminUser{}).Name()
	AdminUserGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: AdminUserKind}.String()
	AdminUserKindAPIVersion   = AdminUserKind + "." + SchemeGroupVersion.String()
	AdminUserGroupVersionKind = SchemeGroupVersion.WithKind(AdminUserKind)
)

func init() {
	SchemeBuilder.Register(&AdminUser{}, &AdminUserList{})
}
