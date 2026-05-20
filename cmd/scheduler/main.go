package main

import (
	"os"

	"k8s.io/component-base/cli"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	_ "k8s.io/component-base/metrics/prometheus/version"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"

	_ "sigs.k8s.io/scheduler-plugins/apis/config/scheme"

	"sigs.k8s.io/scheduler-plugins/pkg/edgefit"
)

func main() {
	command := app.NewSchedulerCommand(
		app.WithPlugin(edgefit.Name, edgefit.New),
	)

	os.Exit(cli.Run(command))
}
