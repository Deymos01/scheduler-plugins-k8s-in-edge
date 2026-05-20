package edgefit

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	"sigs.k8s.io/scheduler-plugins/apis/config"
)

const Name = "EdgeFit"

type EdgeFit struct {
	handle  framework.Handle
	args    *config.EdgeFitArgs
	runtime *runtimeConfig
}

var _ framework.ScorePlugin = &EdgeFit{}

func New(ctx context.Context, obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	args := defaultArgs()

	if obj != nil {
		pluginArgs, ok := obj.(*config.EdgeFitArgs)
		if !ok {
			return nil, fmt.Errorf("unexpected args type %T", obj)
		}
		mergeArgs(args, pluginArgs)
	}

	if err := validateArgs(args); err != nil {
		return nil, err
	}

	runtimeConfig, err := newRuntimeConfig(args)
	if err != nil {
		return nil, err
	}

	klog.FromContext(ctx).Info(
		"created EdgeFit plugin",
		"nodeClassLabel", args.NodeClassLabel,
		"classes", len(runtimeConfig.classes),
		"weightClass", args.Weights.Class,
		"weightPacking", args.Weights.Packing,
		"weightBalance", args.Weights.Balance,
		"weightFragmentation", args.Weights.Fragmentation,
	)

	return &EdgeFit{
		handle:  handle,
		args:    args,
		runtime: runtimeConfig,
	}, nil
}

func (p *EdgeFit) Name() string {
	return Name
}

func (p *EdgeFit) Score(
	ctx context.Context,
	_ *framework.CycleState,
	pod *corev1.Pod,
	nodeInfo *framework.NodeInfo,
) (int64, *framework.Status) {
	input, status := p.collectScoreInput(pod, nodeInfo)
	if status != nil && !status.IsSuccess() {
		return 0, status
	}

	result := p.computeScore(input)

	klog.FromContext(ctx).V(4).Info(
		"EdgeFit score",
		"pod", klog.KObj(pod),
		"node", input.nodeName,
		"nodeClass", input.nodeClassName,
		"minClass", input.minSuitableClassName,
		"classScore", result.classScore,
		"packingScore", result.packingScore,
		"balanceScore", result.balanceScore,
		"fragmentationScore", result.fragmentationScore,
		"score", result.finalScore,
	)

	return result.finalScore, nil
}

func (p *EdgeFit) ScoreExtensions() framework.ScoreExtensions {
	return nil
}
