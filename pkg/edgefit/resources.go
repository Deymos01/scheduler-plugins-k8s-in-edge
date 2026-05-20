package edgefit

import (
	corev1 "k8s.io/api/core/v1"
	resourcehelper "k8s.io/component-helpers/resource"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type scoreInput struct {
	nodeName string

	nodeClassName  string
	nodeClassKnown bool
	nodeClassRank  int32

	minSuitableClassName  string
	minSuitableClassFound bool
	minSuitableClassRank  int32

	podMilliCPU int64
	podMemory   int64

	requestedMilliCPU int64
	requestedMemory   int64

	allocatableMilliCPU int64
	allocatableMemory   int64

	cpuUtilAfter    float64
	memoryUtilAfter float64

	freeMilliCPUAfter int64
	freeMemoryAfter   int64
}

func (p *EdgeFit) collectScoreInput(
	pod *corev1.Pod,
	nodeInfo *framework.NodeInfo,
) (scoreInput, *framework.Status) {
	if pod == nil {
		return scoreInput{}, framework.NewStatus(framework.Error, "pod is nil")
	}
	if nodeInfo == nil {
		return scoreInput{}, framework.NewStatus(framework.Error, "nodeInfo is nil")
	}

	node := nodeInfo.Node()
	if node == nil {
		return scoreInput{}, framework.NewStatus(framework.Error, "node is nil")
	}
	if nodeInfo.Requested == nil {
		return scoreInput{}, framework.NewStatus(framework.Error, "nodeInfo.Requested is nil")
	}
	if nodeInfo.Allocatable == nil {
		return scoreInput{}, framework.NewStatus(framework.Error, "nodeInfo.Allocatable is nil")
	}

	podCPU, podMemory := podRequests(pod)

	reqCPU := nodeInfo.Requested.MilliCPU
	reqMemory := nodeInfo.Requested.Memory

	allocCPU := nodeInfo.Allocatable.MilliCPU
	allocMemory := nodeInfo.Allocatable.Memory

	if allocCPU <= 0 {
		return scoreInput{}, framework.NewStatus(framework.Error, "node allocatable CPU must be positive")
	}
	if allocMemory <= 0 {
		return scoreInput{}, framework.NewStatus(framework.Error, "node allocatable memory must be positive")
	}

	cpuAfter := reqCPU + podCPU
	memoryAfter := reqMemory + podMemory

	nodeClassName, nodeClassKnown, nodeClassRank := p.nodeClass(node.Labels[p.args.NodeClassLabel])
	minClassName, minClassFound, minClassRank := p.minClass(podCPU, podMemory)

	return scoreInput{
		nodeName: node.Name,

		nodeClassName:  nodeClassName,
		nodeClassKnown: nodeClassKnown,
		nodeClassRank:  nodeClassRank,

		minSuitableClassName:  minClassName,
		minSuitableClassFound: minClassFound,
		minSuitableClassRank:  minClassRank,

		podMilliCPU: podCPU,
		podMemory:   podMemory,

		requestedMilliCPU: reqCPU,
		requestedMemory:   reqMemory,

		allocatableMilliCPU: allocCPU,
		allocatableMemory:   allocMemory,

		cpuUtilAfter:    utilization(cpuAfter, allocCPU),
		memoryUtilAfter: utilization(memoryAfter, allocMemory),

		freeMilliCPUAfter: allocCPU - cpuAfter,
		freeMemoryAfter:   allocMemory - memoryAfter,
	}, nil
}

func (p *EdgeFit) nodeClass(name string) (string, bool, int32) {
	class, ok := p.runtime.classByName[name]
	if !ok {
		return name, false, -1
	}

	return name, true, class.rank
}

func (p *EdgeFit) minClass(cpu, memory int64) (string, bool, int32) {
	class, ok := p.runtime.minSuitableClass(cpu, memory)
	if !ok {
		return "", false, -1
	}

	return class.name, true, class.rank
}

func podRequests(pod *corev1.Pod) (int64, int64) {
	requests := resourcehelper.PodRequests(pod, resourcehelper.PodResourcesOptions{})

	cpu := requests[corev1.ResourceCPU]
	memory := requests[corev1.ResourceMemory]

	return cpu.MilliValue(), memory.Value()
}

func utilization(used, capacity int64) float64 {
	if capacity <= 0 {
		return 0
	}

	return float64(used) / float64(capacity)
}
