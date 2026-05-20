package edgefit

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/api/resource"

	"sigs.k8s.io/scheduler-plugins/apis/config"
)

type nodeClassRuntime struct {
	name     string
	milliCPU int64
	memory   int64
	rank     int32
}

type runtimeConfig struct {
	classes             []nodeClassRuntime
	classByName         map[string]nodeClassRuntime
	minFragmentMilliCPU int64
	minFragmentMemory   int64
	weightSum           float64
}

func newRuntimeConfig(args *config.EdgeFitArgs) (*runtimeConfig, error) {
	classes := make([]nodeClassRuntime, 0, len(args.Classes))
	classByName := make(map[string]nodeClassRuntime, len(args.Classes))

	for _, class := range args.Classes {
		runtimeClass, err := parseNodeClass(class)
		if err != nil {
			return nil, err
		}

		classes = append(classes, runtimeClass)
		classByName[runtimeClass.name] = runtimeClass
	}

	sort.Slice(classes, func(i, j int) bool {
		return classes[i].rank < classes[j].rank
	})

	minCPU, err := resource.ParseQuantity(args.MinFragment.CPU)
	if err != nil {
		return nil, fmt.Errorf("parse minFragment.cpu: %w", err)
	}

	minMemory, err := resource.ParseQuantity(args.MinFragment.Memory)
	if err != nil {
		return nil, fmt.Errorf("parse minFragment.memory: %w", err)
	}

	return &runtimeConfig{
		classes:             classes,
		classByName:         classByName,
		minFragmentMilliCPU: minCPU.MilliValue(),
		minFragmentMemory:   minMemory.Value(),
		weightSum:           weightsSum(args.Weights),
	}, nil
}

func parseNodeClass(class config.EdgeFitNodeClass) (nodeClassRuntime, error) {
	cpu, err := resource.ParseQuantity(class.CPU)
	if err != nil {
		return nodeClassRuntime{}, fmt.Errorf("parse cpu for class %q: %w", class.Name, err)
	}

	memory, err := resource.ParseQuantity(class.Memory)
	if err != nil {
		return nodeClassRuntime{}, fmt.Errorf("parse memory for class %q: %w", class.Name, err)
	}

	return nodeClassRuntime{
		name:     class.Name,
		milliCPU: cpu.MilliValue(),
		memory:   memory.Value(),
		rank:     class.Rank,
	}, nil
}

func weightsSum(weights config.EdgeFitWeights) float64 {
	return weights.Class +
		weights.Packing +
		weights.Balance +
		weights.Fragmentation
}

func (r *runtimeConfig) minSuitableClass(podMilliCPU, podMemory int64) (nodeClassRuntime, bool) {
	for _, class := range r.classes {
		if class.milliCPU >= podMilliCPU && class.memory >= podMemory {
			return class, true
		}
	}

	return nodeClassRuntime{}, false
}
