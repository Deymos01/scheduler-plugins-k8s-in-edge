package config

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EdgeFitArgs struct {
	metav1.TypeMeta

	NodeClassLabel string
	Weights        EdgeFitWeights
	Classes        []EdgeFitNodeClass
	MinFragment    EdgeFitResource
}

type EdgeFitWeights struct {
	Class         float64
	Packing       float64
	Balance       float64
	Fragmentation float64
}

type EdgeFitNodeClass struct {
	Name   string
	CPU    string
	Memory string
	Rank   int32
}

type EdgeFitResource struct {
	CPU    string
	Memory string
}
