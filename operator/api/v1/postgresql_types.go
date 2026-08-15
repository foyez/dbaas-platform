package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type PostgreSQLPhase string
type PostgreSQLConditionType string

const (
	PostgreSQLPhasePending  PostgreSQLPhase = "Pending"
	PostgreSQLPhaseCreating PostgreSQLPhase = "Creating"
	PostgreSQLPhaseReady    PostgreSQLPhase = "Ready"
	PostgreSQLPhaseFailed   PostgreSQLPhase = "Failed"
)

const (
	PostgreSQLConditionReady PostgreSQLConditionType = "Ready"
)

type PostgreSQLSpec struct {
	// +kubebuilder:validation:Required
	Version string `json:"version"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	Instances int32 `json:"instances,omitempty"`

	// +kubebuilder:validation:Required
	Storage string `json:"storage"`

	// +kubebuilder:validation:Required
	Database string `json:"database"`

	// +kubebuilder:validation:Required
	Username string `json:"username"`
}

type PostgreSQLStatus struct {
	Phase          PostgreSQLPhase    `json:"phase,omitempty"`
	ReadyInstances int32              `json:"readyInstances,omitempty"`
	Conditions     []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="PHASE",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="READY",type="integer",JSONPath=".status.readyInstances"
type PostgreSQL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgreSQLSpec   `json:"spec"`
	Status PostgreSQLStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type PostgreSQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgreSQL `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(
			SchemeGroupVersion,
			&PostgreSQL{},
			&PostgreSQLList{},
		)

		return nil
	})
}
