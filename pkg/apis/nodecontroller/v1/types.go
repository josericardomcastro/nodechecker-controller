/*
Copyright 2022
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

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="LastSchedule",type=string,JSONPath=`.status.lastSchedule`
// +kubebuilder:printcolumn:name="LastStatus",type=string,JSONPath=`.status.lastStatus`

// NodeChecker is specification for a NodeChecker resource
type NodeChecker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeCheckerSpec   `json:"spec,omitempty"`
	Status NodeCheckerStatus `json:"status,omitempty"`
}

type NodeCheckerSpec struct {
	// +kubebuilder:validation:Enum=connection;feature
	Type            string                     `json:"type"`
	Schedule        string                     `json:"schedule"`
	SourceNodes     NodeCheckerSpecSourceNodes `json:"sourceNodes"`
	ConnectionCheck ConnectionCheck            `json:"connectionCheck,omitempty"`
	FeatureCheck    FeatureCheck               `json:"featureCheck,omitempty"`
}

type NodeCheckerStatus struct {
	LastSchedule string `json:"lastSchedule"`
	LastStatus   string `json:"lastStatus"`
}

type NodeCheckerSpecSourceNodes struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type FeatureCheck struct {
	Command    string            `json:"command,omitempty"`
	SyncLabels map[string]string `json:"syncLabels,omitempty"`
}

type ConnectionCheck struct {
	External         []TargetExternal         `json:"external,omitempty"`
	ClusterNodes     []TargetClusterNodes     `json:"clusterNodes,omitempty"`
	ClusterEndpoints []TargetClusterEndpoints `json:"clusterEndpoints,omitempty"`
}

type TargetExternal struct {
	Name       string            `json:"name"`
	Host       string            `json:"host"`
	Port       int               `json:"port"`
	Protocol   string            `json:"protocol"`
	SyncLabels map[string]string `json:"syncLabels,omitempty"`
}

type TargetClusterNodes struct {
	Name        string            `json:"name"`
	MatchLabels map[string]string `json:"matchLabels"`
	Port        int               `json:"port"`
	Protocol    string            `json:"protocol"`
	SyncLabels  map[string]string `json:"syncLabels,omitempty"`
}

type TargetClusterEndpoints struct {
	Name       string                         `json:"name"`
	Endpoint   TargetClusterEndpointsEndpoint `json:"endpoint"`
	SyncLabels map[string]string              `json:"syncLabels,omitempty"`
}

type TargetClusterEndpointsEndpoint struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	SyncLabels map[string]string `json:"syncLabels,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NodeCheckerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeChecker `json:"items"`
}
