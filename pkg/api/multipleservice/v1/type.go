package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceExport resource
type ServiceExport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ServiceExportSpec `json:"spec,omitempty"`

	Status ServiceExportStatus `json:"status,omitempty"`
}

// ServiceExportSpec describes the attributes that a ServiceExport is created with.
type ServiceExportSpec struct {
	FromLabels       []string   `json:"fromLabels"`
	FromAnnoatations []string   `json:"fromAnnoatations"`
	FromNamespaces   []string   `json:"fromNamespaces"`
	ToCluster        *ToCluster `json:"toCluster"`
}

// ToCluster describes the cluste with id and namespace
type ToCluster struct {
	ClusterID string `json:"clusterID"`
	Namespace string `json:"namespace"`
}

// ServiceExportStatus is information about the current status of a ServiceExport.
type ServiceExportStatus struct {
	// Information when was the last time the schedule was successfully scheduled.
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty" protobuf:"bytes,2,opt,name=lastScheduleTime"`

	Conditions []ServiceExportConditions `json:"conditions"`
}

// ServiceExportConditions contains condition information for a ServiceExport
type ServiceExportConditions struct {
	// TODO: add condition for serviceExport
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceExportList is a set of ServiceExport
type ServiceExportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ServiceExport `json:"items"`
}
