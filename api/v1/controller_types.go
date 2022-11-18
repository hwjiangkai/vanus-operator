/*
Copyright 2022.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ControllerSpec defines the desired state of Controller
type ControllerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Replicas is the number of nodes in the Controller. Each node is deployed as a Replica in a StatefulSet. Only 1, 3, 5 replicas clusters are tested.
	// This value should be an odd number to ensure the resultant cluster can establish exactly one quorum of nodes
	// in the event of a fragmenting network partition.
	// +optional
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:default:=1
	Replicas int32 `json:"replicas,omitempty"`
	// Image is the name of the controller docker image to use for the Pods.
	// Must be provided together with ImagePullSecrets in order to use an image in a private registry.
	Image string `json:"image,omitempty"`
	// ImagePullPolicy defines how the image is pulled
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// List of Secret resource containing access credentials to the registry for the Controller image. Required if the docker registry is private.
	// ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// The desired compute resource requirements of Pods in the cluster.
	// +kubebuilder:default:={limits: {cpu: "2000m", memory: "2Gi"}, requests: {cpu: "1000m", memory: "2Gi"}}
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// StorageMode can be EmptyDir, HostPath, StorageClass
	// StorageMode string `json:"storageMode"`
	// HostPath is the local path to store data
	// HostPath string `json:"hostPath"`
	// Env defines custom env
	// Env []corev1.EnvVar `json:"env,omitempty"`
	// VolumeClaimTemplates defines the StorageClass
	// VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates"`
	// Pod Security Context
	// PodSecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	// Container Security Context
	// ContainerSecurityContext *corev1.SecurityContext `json:"containerSecurityContext,omitempty"`
	// Affinity scheduling rules to be applied on created Pods.
	// Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Tolerations is the list of Toleration resources attached to each Pod in the Controller.
	// Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node
	// NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// PriorityClassName indicates the pod's priority
	// PriorityClassName string `json:"priorityClassName,omitempty"`
	// ServiceAccountName
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// ControllerStatus defines the observed state of Controller
type ControllerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes []string `json:"nodes"`
	Size  int32    `json:"size"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Controller is the Schema for the controllers API
type Controller struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControllerSpec   `json:"spec,omitempty"`
	Status ControllerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ControllerList contains a list of Controller
type ControllerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Controller `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Controller{}, &ControllerList{})
}
