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
type UserGroupBindingParameters struct {
	GroupId         string          `json:"groupId,omitempty"`
	GroupIdRef      *xpv1.Reference `json:"groupIdRef,omitempty"`
	GroupIdSelector *xpv1.Selector  `json:"groupIdSelector,omitempty"`

	UserId         string          `json:"userId,omitempty"`
	UserIdRef      *xpv1.Reference `json:"userIdRef,omitempty"`
	UserIdSelector *xpv1.Selector  `json:"userIdSelector,omitempty"`
}

// UserGroupBindingObservation are the observable fields of a UserGroupBinding.
type UserGroupBindingObservation struct {
	GroupId string `json:"groupId"`
	UserId  string `json:"userId"`
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

// A UserGroupBinding is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
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
	UserGroupBindingGroupKind        = schema.GroupKind{Group: Group, Kind: UserGroupBindingKind}.String()
	UserGroupBindingKindAPIVersion   = UserGroupBindingKind + "." + SchemeGroupVersion.String()
	UserGroupBindingGroupVersionKind = SchemeGroupVersion.WithKind(UserGroupBindingKind)
)

func init() {
	SchemeBuilder.Register(&UserGroupBinding{}, &UserGroupBindingList{})
}
