package edgefit

import (
	"math"

	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type scoreResult struct {
	classScore         float64
	packingScore       float64
	balanceScore       float64
	fragmentationScore float64
	weightedScore      float64
	finalScore         int64
}

func (p *EdgeFit) computeScore(input scoreInput) scoreResult {
	classScore := computeClassScore(input)
	packingScore := computePackingScore(input)
	balanceScore := computeBalanceScore(input)
	fragmentationScore := p.computeFragmentationScore(input)

	weights := p.args.Weights
	weightSum := p.runtime.weightSum
	if weightSum <= 0 {
		weightSum = 1
	}

	score := weights.Class*classScore +
		weights.Packing*packingScore +
		weights.Balance*balanceScore +
		weights.Fragmentation*fragmentationScore

	weightedScore := score / weightSum

	return scoreResult{
		classScore:         classScore,
		packingScore:       packingScore,
		balanceScore:       balanceScore,
		fragmentationScore: fragmentationScore,
		weightedScore:      weightedScore,
		finalScore:         roundScore(weightedScore),
	}
}

func computeClassScore(input scoreInput) float64 {
	if !input.nodeClassKnown || !input.minSuitableClassFound {
		return 0
	}

	distance := input.nodeClassRank - input.minSuitableClassRank
	if distance < 0 {
		return 0
	}

	return clamp(100-25*float64(distance), 0, 100)
}

func computePackingScore(input scoreInput) float64 {
	cpu := input.cpuUtilAfter
	memory := input.memoryUtilAfter

	return clamp(100*(cpu+memory)/2, 0, 100)
}

func computeBalanceScore(input scoreInput) float64 {
	diff := math.Abs(input.cpuUtilAfter - input.memoryUtilAfter)

	return clamp(100*(1-diff), 0, 100)
}

func (p *EdgeFit) computeFragmentationScore(input scoreInput) float64 {
	freeCPU := input.freeMilliCPUAfter
	freeMemory := input.freeMemoryAfter

	minCPU := p.runtime.minFragmentMilliCPU
	minMemory := p.runtime.minFragmentMemory

	if freeCPU < 0 || freeMemory < 0 {
		return 0
	}

	if minCPU <= 0 || minMemory <= 0 {
		return 100
	}

	if freeCPU < minCPU && freeMemory < minMemory {
		return 100
	}

	cpuRatio := float64(freeCPU) / float64(minCPU)
	memoryRatio := float64(freeMemory) / float64(minMemory)

	return clamp(100*math.Min(1, math.Min(cpuRatio, memoryRatio)), 0, 100)
}

func roundScore(value float64) int64 {
	value = clamp(value, 0, float64(framework.MaxNodeScore))

	return int64(math.Round(value))
}

func clamp(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}

	if value > maxValue {
		return maxValue
	}

	return value
}
