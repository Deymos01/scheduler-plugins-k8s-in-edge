package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EdgeFitArgs struct {
	metav1.TypeMeta `json:",inline"`

	NodeClassLabel string             `json:"nodeClassLabel,omitempty"`
	Weights        EdgeFitWeights     `json:"weights,omitempty"`
	Classes        []EdgeFitNodeClass `json:"classes,omitempty"`
	MinFragment    EdgeFitResource    `json:"minFragment,omitempty"`
}

type EdgeFitWeights struct {
	Class         float64 `json:"class,omitempty"`
	Packing       float64 `json:"packing,omitempty"`
	Balance       float64 `json:"balance,omitempty"`
	Fragmentation float64 `json:"fragmentation,omitempty"`
}

type EdgeFitNodeClass struct {
	Name   string `json:"name,omitempty"`
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
	Rank   int32  `json:"rank,omitempty"`
}

type EdgeFitResource struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}
