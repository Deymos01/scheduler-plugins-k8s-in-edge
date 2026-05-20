package edgefit

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"

	"sigs.k8s.io/scheduler-plugins/apis/config"
)

func defaultArgs() *config.EdgeFitArgs {
	return &config.EdgeFitArgs{
		NodeClassLabel: "node-class",
		Weights: config.EdgeFitWeights{
			Class:         0.30,
			Packing:       0.25,
			Balance:       0.25,
			Fragmentation: 0.20,
		},
		Classes: []config.EdgeFitNodeClass{
			{Name: "iot-c", CPU: "1", Memory: "2Gi", Rank: 0},
			{Name: "iot-b", CPU: "2", Memory: "3Gi", Rank: 1},
			{Name: "iot-a", CPU: "2", Memory: "4Gi", Rank: 2},
			{Name: "gateway", CPU: "4", Memory: "16Gi", Rank: 3},
			{Name: "powerful", CPU: "8", Memory: "32Gi", Rank: 4},
		},
		MinFragment: config.EdgeFitResource{
			CPU:    "500m",
			Memory: "500Mi",
		},
	}
}

func mergeArgs(dst, src *config.EdgeFitArgs) {
	if src.NodeClassLabel != "" {
		dst.NodeClassLabel = src.NodeClassLabel
	}

	if src.Weights.Class != 0 ||
		src.Weights.Packing != 0 ||
		src.Weights.Balance != 0 ||
		src.Weights.Fragmentation != 0 {
		dst.Weights = src.Weights
	}

	if len(src.Classes) > 0 {
		dst.Classes = src.Classes
	}

	if src.MinFragment.CPU != "" {
		dst.MinFragment.CPU = src.MinFragment.CPU
	}

	if src.MinFragment.Memory != "" {
		dst.MinFragment.Memory = src.MinFragment.Memory
	}
}

func validateArgs(args *config.EdgeFitArgs) error {
	if args.NodeClassLabel == "" {
		return fmt.Errorf("nodeClassLabel must not be empty")
	}

	if len(args.Classes) == 0 {
		return fmt.Errorf("classes must not be empty")
	}

	if args.Weights.Class < 0 ||
		args.Weights.Packing < 0 ||
		args.Weights.Balance < 0 ||
		args.Weights.Fragmentation < 0 {
		return fmt.Errorf("weights must be non-negative")
	}

	weightSum := args.Weights.Class +
		args.Weights.Packing +
		args.Weights.Balance +
		args.Weights.Fragmentation

	if weightSum <= 0 {
		return fmt.Errorf("sum of weights must be positive")
	}

	minCPU, err := resource.ParseQuantity(args.MinFragment.CPU)
	if err != nil {
		return fmt.Errorf("invalid minFragment.cpu %q: %w", args.MinFragment.CPU, err)
	}

	if minCPU.MilliValue() <= 0 {
		return fmt.Errorf("minFragment.cpu must be positive")
	}

	minMemory, err := resource.ParseQuantity(args.MinFragment.Memory)
	if err != nil {
		return fmt.Errorf("invalid minFragment.memory %q: %w", args.MinFragment.Memory, err)
	}

	if minMemory.Value() <= 0 {
		return fmt.Errorf("minFragment.memory must be positive")
	}

	seenRanks := map[int32]string{}
	seenNames := map[string]struct{}{}

	for _, class := range args.Classes {
		if class.Name == "" {
			return fmt.Errorf("node class name must not be empty")
		}

		if _, exists := seenNames[class.Name]; exists {
			return fmt.Errorf("duplicate node class name %q", class.Name)
		}
		seenNames[class.Name] = struct{}{}

		if prev, exists := seenRanks[class.Rank]; exists {
			return fmt.Errorf("duplicate node class rank %d for %q and %q", class.Rank, prev, class.Name)
		}
		seenRanks[class.Rank] = class.Name

		classCPU, err := resource.ParseQuantity(class.CPU)
		if err != nil {
			return fmt.Errorf("invalid cpu for class %q: %q: %w", class.Name, class.CPU, err)
		}

		if classCPU.MilliValue() <= 0 {
			return fmt.Errorf("cpu for class %q must be positive", class.Name)
		}

		classMemory, err := resource.ParseQuantity(class.Memory)
		if err != nil {
			return fmt.Errorf("invalid memory for class %q: %q: %w", class.Name, class.Memory, err)
		}

		if classMemory.Value() <= 0 {
			return fmt.Errorf("memory for class %q must be positive", class.Name)
		}
	}

	return nil
}
