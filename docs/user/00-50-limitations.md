# Serverless Limitations

## Controller Limitations

Function Ccontroller does not serve time-critical requests from users.
It reconciles Function custom resources (CR), stored at the Kubernetes API Server, and has no persistent state on its own.

Function Controller doesn't serve Functions using its allocated runtime resources. It delegates this work to the dedicated Kubernetes workloads.
Refer to the [architecture](technical-reference/04-10-architecture.md) diagram for more details.

Having this in mind, also remember that Function Controller does not require horizontal scaling.
It scales vertically up to `160Mi` of memory and `500m` of CPU time.

## Namespace Setup Limitations

Be aware that if you apply [LimitRanges](https://kubernetes.io/docs/concepts/policy/limit-range/) in the target namespace where you create Functions, the limits also apply to the Function workloads and may prevent Functions from being run. In such cases, ensure that resources requested in the Function configuration are lower than the limits applied in the namespace.

## Limitation for the Number of Functions

There is no upper limit of Functions that you can run on Kyma. Once you define a Function and runtime Pods are always requested by Function Controller. It's up to Kubernetes to schedule them based on the available memory and CPU time on the Kubernetes worker nodes. This is determined mainly by the number of the Kubernetes worker nodes (and the node auto-scaling capabilities) and their computational capacity.

## Runtime Phase Limitations

> [!NOTE]
> All measurements were taken on Kubernetes with five AWS worker nodes of type m5.xlarge (four CPU 3.1 GHz x86_64 cores, 16 GiB memory).

Functions serve user-provided logic wrapped in the web framework, Express for Node.js and Bottle for Python. Taking the user logic aside, those frameworks have limitations and depend on the selected [runtime profile](technical-reference/07-80-available-presets.md#functions-resources) and the Kubernetes nodes specification.

The following tables present the response times of the selected runtime profiles for a "Hello World" Function requested at 50 requests/second. This describes the overhead of the serving framework itself. Any user logic added on top of that adds extra milliseconds and must be profiled separately.

<!-- tabs:start -->

#### **Node.js**

|                               | XL     | L      | M      | S      | XS      |
|-------------------------------|--------|--------|--------|--------|---------|
| response time [average]       | ~13ms  | 13ms   | ~15ms  | ~60ms  | ~400ms  |
| response time [95 percentile] | ~20ms  | ~30ms  | ~70ms  | ~200ms | ~800ms  |
| response time [99 percentile] | ~200ms | ~200ms | ~220ms | ~500ms | ~1.25ms |

#### **Python**

|                               | XL     | L      | M      | S      |
|-------------------------------|--------|--------|--------|--------|
| response time [average]       | ~11ms  | 12ms   | ~12ms  | ~14ms  |
| response time [95 percentile] | ~25ms  | ~25ms  | ~25ms  | ~25ms  |
| response time [99 percentile] | ~175ms | ~180ms | ~210ms | ~280ms |

<!-- tabs:end -->

The bigger the runtime profile, the more resources are available to serve the response quicker. Consider these limits of the serving layer as a baseline because this does not take your Function logic into account.

### Scaling

Function runtime Pods can be scaled horizontally from zero up to the limits of the available resources at the Kubernetes worker nodes.
See the [Use External Scalers](tutorials/01-130-use-external-scalers.md) tutorial for more information.

