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

// ClientGroupBindingParameters are the configurable fields of a ClientGroupBinding.
type ClientGroupBindingParameters struct {
	ConfigurableField string `json:"configurableField"`
}

// ClientGroupBindingObservation are the observable fields of a ClientGroupBinding.
type ClientGroupBindingObservation struct {
	ConfigurableField string `json:"configurableField"`
	ObservableField   string `json:"observableField,omitempty"`
}

// A ClientGroupBindingSpec defines the desired state of a ClientGroupBinding.
type ClientGroupBindingSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ClientGroupBindingParameters `json:"forProvider"`
}

// A ClientGroupBindingStatus represents the observed state of a ClientGroupBinding.
type ClientGroupBindingStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ClientGroupBindingObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A ClientGroupBinding is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,pocketid}
type ClientGroupBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClientGroupBindingSpec   `json:"spec"`
	Status ClientGroupBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClientGroupBindingList contains a list of ClientGroupBinding
type ClientGroupBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClientGroupBinding `json:"items"`
}

// ClientGroupBinding type metadata.
var (
	ClientGroupBindingKind             = reflect.TypeOf(ClientGroupBinding{}).Name()
	ClientGroupBindingGroupKind        = schema.GroupKind{Group: Group, Kind: ClientGroupBindingKind}.String()
	ClientGroupBindingKindAPIVersion   = ClientGroupBindingKind + "." + SchemeGroupVersion.String()
	ClientGroupBindingGroupVersionKind = SchemeGroupVersion.WithKind(ClientGroupBindingKind)
)

func init() {
	SchemeBuilder.Register(&ClientGroupBinding{}, &ClientGroupBindingList{})
}
