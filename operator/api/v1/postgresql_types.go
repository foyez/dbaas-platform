package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PostgreSQLSpec defines the desired state of PostgreSQL
type PostgreSQLSpec struct {
	// Version specifies the PostgreSQL major version (for example, 16 or 17).
	// +kubebuilder:validation:Required
	Version string `json:"version"`

	// Instances is the number of PostgreSQL instances in the cluster.
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	Instances int32 `json:"instances,omitempty"`

	// Storage is the persistent volume size for each PostgreSQL instance.
	// +kubebuilder:validation:Required
	Storage string `json:"storage"`

	// Database is the name of the initial database to create.
	// +kubebuilder:validation:Required
	Database string `json:"database"`

	// Username is the owner of the initial database.
	// +kubebuilder:validation:Required
	Username string `json:"username"`

	// ConnectionPooler configures the optional PgBouncer connection pooler.
	// ConnectionPooler ConnectionPoolerSpec `json:"connectionPooler,omitempty"`
}

type ConnectionPoolerSpec struct {
	// Enabled enables deployment of a PgBouncer connection pooler.
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// PoolSize is the maximum number of backend connections.
	// +kubebuilder:default=20
	// +kubebuilder:validation:Minimum=1
	PoolSize int32 `json:"poolSize,omitempty"`
}

// PostgreSQLStatus defines the observed state of PostgreSQL.
type PostgreSQLStatus struct {
	Phase          string `json:"phase,omitempty"`
	ReadyInstances int32  `json:"readyInstances,omitempty"`
	// ConnectionEndpoint string `json:"connectionEndpoint,omitempty"`
	// PgBouncerStatus    string `json:"pgBouncerStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="PHASE",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="READY",type="integer",JSONPath=".status.readyInstances"
// PostgreSQL is the Schema for the postgresqls API
type PostgreSQL struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of PostgreSQL
	// +required
	Spec PostgreSQLSpec `json:"spec"`

	// status defines the observed state of PostgreSQL
	// +optional
	Status PostgreSQLStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PostgreSQLList contains a list of PostgreSQL
type PostgreSQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PostgreSQL `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &PostgreSQL{}, &PostgreSQLList{})
		return nil
	})
}
