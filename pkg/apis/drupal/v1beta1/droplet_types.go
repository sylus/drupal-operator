/*

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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecretRef represents a reference to a Secret
type SecretRef string

// Domain represents a valid domain name
type Domain string

// DropletSpec defines the desired state of Droplet
type DropletSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Domains for which this this site answers.
	// The first item is set as the "main domain" (eg. DRUPAL_HOME and DRUPAL_SITEURL constants).
	// +kubebuilder:validation:MinItems=1
	Domains []Domain `json:"domains"`
	// DrupalSpec for related configuration overrides
	// +optional
	Drupal DrupalSpec `json:"drupal,omitempty"`
	// NginxSpec for related configuration overrides
	// +optional
	Nginx NginxSpec `json:"nginx,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run this
	// site's pods
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// TLSSecretRef a secret containing the TLS certificates for this site.
	// +optional
	TLSSecretRef SecretRef `json:"tlsSecretRef,omitempty"`
	// IngressAnnotations for this Droplet site
	// +optional
	IngressAnnotations map[string]string `json:"ingressAnnotations,omitempty"`
}

// DrupalSpec desired configuration for Drupal
type DrupalSpec struct {
	// Number of desired web pods. This is a pointer to distinguish between
	// explicit zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Drupal runtime image to use. Defaults to drupalwxt/site-canada
	// +optional
	Image string `json:"image,omitempty"`
	// Image tag to use. Defaults to latest
	// +optional
	Tag string `json:"tag,omitempty"`
	// ImagePullPolicy overrides DropletRuntime spec.imagePullPolicy
	// +kubebuilder:validation:Enum=Always,IfNotPresent,Never
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// ImagePullSecrets defines additional secrets to use when pulling images
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// CodeVolumeSpec specifies how the site's code gets mounted into the
	// container. If not specified, a code volume won't get mounted at all.
	// +optional
	CodeVolumeSpec *CodeVolumeSpec `json:"code,omitempty"`
	// MediaVolumeSpec specifies how media files get mounted into the runtime
	// container. If not specified, a media volume won't be mounted at all.
	// +optional
	MediaVolumeSpec *MediaVolumeSpec `json:"media,omitempty"`
	// Volumes defines additional volumes to get injected into Drupal pods
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
	// VolumeMountsSpec defines additional mounts which get injected into web
	// and cli pods.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// Env defines environment variables which get passed into Drupal pods
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	// EnvFrom defines envFrom's which get passed into Drupal pods
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
}

// NginxSpec desired configuration for Nginx
type NginxSpec struct {
	// Number of desired web pods. This is a pointer to distinguish between
	// explicit zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Nginx runtime image to use. Defaults to drupalwxt/site-canada
	// +optional
	Image string `json:"image,omitempty"`
	// Image tag to use. Defaults to latest
	// +optional
	Tag string `json:"tag,omitempty"`
	// ImagePullPolicy overrides DropletRuntime spec.imagePullPolicy
	// +kubebuilder:validation:Enum=Always,IfNotPresent,Never
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// ImagePullSecrets defines additional secrets to use when pulling images
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// Volumes defines additional volumes to get injected into Nginx pods
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
	// VolumeMountsSpec defines additional mounts which get injected into web
	// and cli pods.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// Env defines environment variables which get passed into Nginx pods
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	// EnvFrom defines envFrom's which get passed into Nginx pods
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
}

// GitVolumeSource is the desired spec for git code source
type GitVolumeSource struct {
	// Repository is the git repository for the code
	Repository string `json:"repository"`
	// GitRef to clone (can be a branch name, but it should point to a tag or a
	// commit hash)
	// +optional
	GitRef string `json:"reference,omitempty"`
	// Env defines env variables  which get passed to the git clone container
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	// EnvFrom defines envFrom which get passed to the git clone container
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
	// EmptyDir volume to use for git cloning.
	// +optional
	EmptyDir *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
}

// S3VolumeSource is the desired spec for accessing media files over S3
// compatible object store
type S3VolumeSource struct {
	// Bucket for storing media files
	// +kubebuilder:validation:MinLength=1
	Bucket string `json:"bucket"`
	// PathPrefix is the prefix for media files in bucket
	PathPrefix string `json:"prefix,omitempty"`
	// Env variables for accessing S3 bucket. Taken into account are:
	// ACCESS_KEY, SECRET_ACCESS_KEY
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// GCSVolumeSource is the desired spec for accessing media files using google
// cloud storage object store
type GCSVolumeSource struct {
	// Bucket for storing media files
	// +kubebuilder:validation:MinLength=1
	Bucket string `json:"bucket"`
	// PathPrefix is the prefix for media files in bucket
	PathPrefix string `json:"prefix,omitempty"`
	// Env variables for accessing gcs bucket. Taken into account are:
	// GOOGLE_APPLICATION_CREDENTIALS_JSON
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// CodeVolumeSpec is the desired spec for mounting code into the drupal
// runtime container
type CodeVolumeSpec struct {
	// ReadOnly specifies if the volume should be mounted read-only inside the
	// drupal runtime container
	ReadOnly bool
	// MountPath spechfies where should the code volume be mounted.
	// Defaults to /var/www/html/modules/custom
	// +optional
	MountPath string `json:"mountPath,omitempty"`
	// ContentSubPath specifies where within the code volumes, the droplet-content
	// folder resides.
	// Defaults to droplet-content/
	// +optional
	ContentSubPath string `json:"contentSubPath,omitempty"`
	// GitDir specifies the git repo to use for code cloning. It has the highest
	// level of precedence over EmptyDir, HostPath and PersistentVolumeClaim
	// +optional
	GitDir *GitVolumeSource `json:"git,omitempty"`
	// PersistentVolumeClaim to use if no GitDir is specified
	// +optional
	PersistentVolumeClaim *corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
	// HostPath to use if no PersistentVolumeClaim is specified
	// +optional
	HostPath *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`
	// EmptyDir to use if no HostPath is specified
	// +optional
	EmptyDir *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
}

// MediaVolumeSpec is the desired spec for mounting code into the drupal
// runtime container
type MediaVolumeSpec struct {
	// ReadOnly specifies if the volume should be mounted read-only inside the
	// drupal runtime container
	ReadOnly bool
	// S3VolumeSource specifies the S3 object storage configuration for media
	// files. It has the highest level of precedence over EmptyDir, HostPath
	// and PersistentVolumeClaim
	// +optional
	S3VolumeSource *S3VolumeSource `json:"s3,omitempty"`
	// GCSVolumeSource specifies the google cloud storage object storage
	// configuration for media files. It has the highest level of precedence
	// over EmptyDir, HostPath and PersistentVolumeClaim
	// +optional
	GCSVolumeSource *GCSVolumeSource `json:"gcs,omitempty"`
	// PersistentVolumeClaim to use if no S3VolumeSource or GCSVolumeSource are
	// specified
	// +optional
	PersistentVolumeClaim *corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
	// HostPath to use if no PersistentVolumeClaim is specified
	// +optional
	HostPath *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`
	// EmptyDir to use if no HostPath is specified
	// +optional
	EmptyDir *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
}

// DropletStatus defines the observed state of Droplet
type DropletStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Total number of non-terminated pods targeted by web deployment
	// This is copied over from the deployment object
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Droplet is the Schema for the droplets API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
type Droplet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DropletSpec   `json:"spec,omitempty"`
	Status DropletStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DropletList contains a list of Droplet
type DropletList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Droplet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Droplet{}, &DropletList{})
}
