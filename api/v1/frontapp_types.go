/*
Copyright 2023.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FrontAppSpec defines the desired state of FrontApp
type FrontAppSpec struct {
	// image address
	Image string `json:"image,omitempty"`
	// field reverse-proxy for Caddyfile, to specify the api address for front end
	ReverseProxy string `json:"reverseProxy,omitempty"`
	// url of front app
	Url string `json:"url,omitempty"`
}

// FrontAppStatus defines the observed state of FrontApp
type FrontAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// FrontApp is the Schema for the frontapps API
type FrontApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FrontAppSpec   `json:"spec,omitempty"`
	Status FrontAppStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FrontAppList contains a list of FrontApp
type FrontAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FrontApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FrontApp{}, &FrontAppList{})
}
