# EdgeFit Scheduler Plugin

EdgeFit is a custom Kubernetes scheduler plugin developed for scheduling Pods on heterogeneous edge nodes.

The plugin is implemented as an out-of-tree `Score` plugin for `kube-scheduler` using the Kubernetes Scheduling Framework.

## Purpose

EdgeFit is intended for a simulated heterogeneous edge cluster where nodes have different CPU and memory capacities.

The main idea is to place small Pods on smaller suitable nodes, avoid using large nodes too early, and reduce CPU/memory disbalance and fragmentation.

## Node classes

EdgeFit expects nodes to have a class label.

The default node classes are:

| Class | CPU | Memory | Rank |
|---|---:|---:|---:|
| `iot-c` | 1 CPU | 2Gi | 0 |
| `iot-b` | 2 CPU | 3Gi | 1 |
| `iot-a` | 2 CPU | 4Gi | 2 |
| `gateway` | 4 CPU | 16Gi | 3 |
| `powerful` | 8 CPU | 32Gi | 4 |

## Scoring formula

For each feasible node, EdgeFit calculates a score from several components:

```text
score(node, pod) =
    w_class * class_score
  + w_packing * packing_score
  + w_balance * balance_score
  + w_fragmentation * fragmentation_score
```

The result is normalized by the sum of weights and returned as a scheduler score from `0` to `100`.

The default weights are:

```yaml
weights:
  class: 0.30
  packing: 0.25
  balance: 0.25
  fragmentation: 0.20
```

## Score components

### class_score

`class_score` prefers the smallest suitable node class for a Pod.

### packing_score

`packing_score` gives a higher score to nodes that become more utilized after placing the Pod.

### balance_score

`balance_score` penalizes CPU/memory imbalance after placing the Pod.

### fragmentation_score

`fragmentation_score` penalizes resource leftovers that are hard to use for future Pods.

If after placing a Pod the node has a lot of one resource but too little of another resource, the score is reduced.

## Example scheduler configuration

A ready-to-use example configuration is available in:

```text
configs/edgefit-scheduler.yaml
```

Before running it, replace the kubeconfig path with the local path on your machine.

## Build

From the repository root:

```bash
go build -o bin/kube-scheduler ./cmd/scheduler
```

## Run

Start the custom scheduler:

```bash
./bin/kube-scheduler \
  --config configs/edgefit-scheduler.yaml \
  --v=2
```

Pods should use the scheduler profile:

```yaml
spec:
  schedulerName: edge-aware-scheduler
```
