// Package operator contains the Go type definitions for the Pranor Vault
// Custom Resource Definitions (CRDs). These types implement runtime.Object
// and are registered with the controller-runtime scheme.
package main

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ─── PranorVaultCluster ───────────────────────────────────────────────────────

// PranorVaultClusterSpec defines the desired state of a Pranor Vault cluster.
type PranorVaultClusterSpec struct {
	// Replicas is the number of storage nodes to deploy (1–32).
	Replicas int32 `json:"replicas"`
	// Image is the container image for the Pranor Vault server.
	Image string `json:"image"`
	// DataDir is the mount path for object data inside the container.
	// +optional
	DataDir string `json:"dataDir,omitempty"`
	// Auth configures S3 authentication.
	// +optional
	Auth *ClusterAuth `json:"auth,omitempty"`
	// TLS configures TLS 1.3 for the S3 endpoint.
	// +optional
	TLS *ClusterTLS `json:"tls,omitempty"`
	// ErasureCoding enables Reed-Solomon erasure coding.
	// +optional
	ErasureCoding *ErasureCodingConfig `json:"erasureCoding,omitempty"`
	// ReplicationFactor is the number of data replicas (used when erasure coding is disabled).
	// +optional
	ReplicationFactor int32 `json:"replicationFactor,omitempty"`
	// RateLimit configures per-tenant request rate limiting.
	// +optional
	RateLimit *RateLimitConfig `json:"rateLimit,omitempty"`
	// Storage configures the PersistentVolumeClaim for each node.
	// +optional
	Storage *StorageConfig `json:"storage,omitempty"`
}

type ClusterAuth struct {
	Enabled              bool   `json:"enabled"`
	AccessKey            string `json:"accessKey,omitempty"`
	SecretKeySecretRef   *SecretKeyRef `json:"secretKeySecretRef,omitempty"`
}

type SecretKeyRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type ClusterTLS struct {
	SecretName string `json:"secretName"`
}

type ErasureCodingConfig struct {
	Enabled      bool  `json:"enabled"`
	DataShards   int32 `json:"dataShards,omitempty"`
	ParityShards int32 `json:"parityShards,omitempty"`
}

type RateLimitConfig struct {
	RequestsPerSecond int `json:"requestsPerSecond"`
	Burst             int `json:"burst,omitempty"`
}

type StorageConfig struct {
	StorageClassName string `json:"storageClassName,omitempty"`
	Size             string `json:"size,omitempty"`
}

// ClusterPhase is the lifecycle state of a PranorVaultCluster.
type ClusterPhase string

const (
	ClusterPhasePending      ClusterPhase = "Pending"
	ClusterPhaseInitializing ClusterPhase = "Initializing"
	ClusterPhaseRunning      ClusterPhase = "Running"
	ClusterPhaseUpgrading    ClusterPhase = "Upgrading"
	ClusterPhaseDegraded     ClusterPhase = "Degraded"
	ClusterPhaseFailed       ClusterPhase = "Failed"
)

// PranorVaultClusterStatus reflects the observed state of a PranorVaultCluster.
type PranorVaultClusterStatus struct {
	Phase      ClusterPhase       `json:"phase,omitempty"`
	ReadyNodes int32              `json:"readyNodes,omitempty"`
	Endpoint   string             `json:"endpoint,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PranorVaultCluster is the Schema for the pranorVaultclusters API.
type PranorVaultCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PranorVaultClusterSpec   `json:"spec,omitempty"`
	Status            PranorVaultClusterStatus `json:"status,omitempty"`
}

func (s *PranorVaultCluster) DeepCopyObject() runtime.Object {
	out := new(PranorVaultCluster)
	*out = *s
	out.Spec = s.Spec
	out.Status = s.Status
	return out
}

// PranorVaultClusterList is a list of PranorVaultCluster resources.
type PranorVaultClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PranorVaultCluster `json:"items"`
}

func (s *PranorVaultClusterList) DeepCopyObject() runtime.Object {
	out := new(PranorVaultClusterList)
	*out = *s
	out.Items = make([]PranorVaultCluster, len(s.Items))
	copy(out.Items, s.Items)
	return out
}

// ─── PranorVaultBucket ────────────────────────────────────────────────────────

// PranorVaultBucketSpec defines the desired state of a managed bucket.
type PranorVaultBucketSpec struct {
	ClusterRef         string                `json:"clusterRef"`
	Versioning         string                `json:"versioning,omitempty"`
	ContentAddressable bool                  `json:"contentAddressable,omitempty"`
	DeletionPolicy     string                `json:"deletionPolicy,omitempty"`
	Lifecycle          []BucketLifecycleRule `json:"lifecycle,omitempty"`
	ColdTier           *BucketColdTier       `json:"coldTier,omitempty"`
}

type BucketLifecycleRule struct {
	ID             string `json:"id"`
	Enabled        bool   `json:"enabled"`
	Prefix         string `json:"prefix,omitempty"`
	ExpirationDays int    `json:"expirationDays"`
}

type BucketColdTier struct {
	Endpoint        string        `json:"endpoint"`
	RemoteBucket    string        `json:"remoteBucket"`
	Region          string        `json:"region,omitempty"`
	SecretRef       *SecretKeyRef `json:"secretRef,omitempty"`
	MinAgeDays      int           `json:"minAgeDays,omitempty"`
	ScanIntervalMin int           `json:"scanIntervalMin,omitempty"`
}

// BucketPhase is the lifecycle state of a PranorVaultBucket.
type BucketPhase string

const (
	BucketPhasePending BucketPhase = "Pending"
	BucketPhaseReady   BucketPhase = "Ready"
	BucketPhaseFailed  BucketPhase = "Failed"
)

// PranorVaultBucketStatus reflects the observed state of a PranorVaultBucket.
type PranorVaultBucketStatus struct {
	Phase    BucketPhase `json:"phase,omitempty"`
	Endpoint string      `json:"endpoint,omitempty"`
	Message  string      `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PranorVaultBucket is the Schema for the pranorVaultbuckets API.
type PranorVaultBucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PranorVaultBucketSpec   `json:"spec,omitempty"`
	Status            PranorVaultBucketStatus `json:"status,omitempty"`
}

func (s *PranorVaultBucket) DeepCopyObject() runtime.Object {
	out := new(PranorVaultBucket)
	*out = *s
	return out
}

// PranorVaultBucketList is a list of PranorVaultBucket resources.
type PranorVaultBucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PranorVaultBucket `json:"items"`
}

func (s *PranorVaultBucketList) DeepCopyObject() runtime.Object {
	out := new(PranorVaultBucketList)
	*out = *s
	out.Items = make([]PranorVaultBucket, len(s.Items))
	copy(out.Items, s.Items)
	return out
}

// ─── PranorVaultCredential ────────────────────────────────────────────────────

// PranorVaultCredentialSpec defines the desired state of a managed credential.
type PranorVaultCredentialSpec struct {
	ClusterRef string `json:"clusterRef"`
	SecretRef  string `json:"secretRef"`
	Username   string `json:"username,omitempty"`
	Policy     string `json:"policy,omitempty"`
}

// CredentialPhase is the sync state of a PranorVaultCredential.
type CredentialPhase string

const (
	CredentialPhasePending CredentialPhase = "Pending"
	CredentialPhaseSynced  CredentialPhase = "Synced"
	CredentialPhaseFailed  CredentialPhase = "Failed"
)

// PranorVaultCredentialStatus reflects the sync state.
type PranorVaultCredentialStatus struct {
	Phase        CredentialPhase `json:"phase,omitempty"`
	LastSyncTime *metav1.Time    `json:"lastSyncTime,omitempty"`
	Message      string          `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PranorVaultCredential is the Schema for the pranorVaultcredentials API.
type PranorVaultCredential struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PranorVaultCredentialSpec   `json:"spec,omitempty"`
	Status            PranorVaultCredentialStatus `json:"status,omitempty"`
}

func (s *PranorVaultCredential) DeepCopyObject() runtime.Object {
	out := new(PranorVaultCredential)
	*out = *s
	return out
}

// PranorVaultCredentialList is a list of PranorVaultCredential resources.
type PranorVaultCredentialList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PranorVaultCredential `json:"items"`
}

func (s *PranorVaultCredentialList) DeepCopyObject() runtime.Object {
	out := new(PranorVaultCredentialList)
	*out = *s
	out.Items = make([]PranorVaultCredential, len(s.Items))
	copy(out.Items, s.Items)
	return out
}
