# Serverless Limitations

## Controller Limitations

Function Controller does not serve time-critical requests from users.
It reconciles Function custom resources (CR), stored at the Kubernetes API Server, and has no persistent state on its own.

Function Controller doesn't serve Functions using its allocated runtime resources. It delegates this work to the dedicated Kubernetes workloads.
Refer to the [architecture](technical-reference/04-10-architecture.md) diagram for more details.

Having this in mind, also remember that Function Controller does not require horizontal scaling.
It scales vertically up to `1Gi` of memory and `500m` of CPU time.

## Namespace Setup Limitations

Be aware that if you apply [LimitRanges](https://kubernetes.io/docs/concepts/policy/limit-range/) in the target namespace where you create Functions, the limits also apply to the Function workloads and may prevent Functions from being run. In such cases, ensure that resources requested in the Function configuration are lower than the limits applied in the namespace.

## Limitation for the Number of Functions

There is no upper limit of Functions that you can run on Kyma. Once you define a Function, Pods are always requested by Function Controller. It's up to Kubernetes to schedule them based on the available memory and CPU time on the Kubernetes worker nodes. This is determined mainly by the number of the Kubernetes worker nodes (and the node auto-scaling capabilities) and their computational capacity.

## Runtime Phase Limitations

> [!NOTE]
> All measurements were taken on Kubernetes with three Azure worker nodes of type Standard_D2s_v5 (two vCPU amd64 cores, ~8 GiB memory), distributed across availability zones westeurope-1, westeurope-2, and westeurope-3, running Garden Linux 1877.10 with kernel 6.12.66-cloud-amd64 and Kubernetes v1.34.3.
>
> The values in the tables below are averages from three test runs. Last updated: 2026-04-02.

Functions serve user-provided logic wrapped in the web framework, Express for Node.js and Bottle for Python. Taking the user logic aside, those frameworks have limitations and depend on the selected [runtime profile](technical-reference/07-80-available-presets.md#functions-resources) and the Kubernetes nodes specification.

The following tables present the response times of the selected runtime profiles for a "Hello World" Function across three load scenarios. This describes the overhead of the serving framework itself. Any user logic added on top of that adds extra milliseconds and must be profiled separately.

Tests are implemented using [k6](https://k6.io/) and consist of the following scenarios:

- **Constant load** — 50 virtual users send one request per second each (with a 1-second sleep between calls) for 2 minutes. Represents a steady, moderate traffic baseline.
- **Max load** — 100 virtual users send requests as fast as possible (no sleep) for 2 minutes. Represents sustained high concurrency.

### Constant load

#### Node.js 22

| response time [ms]        |  XS |   S |   M |   L |  XL |
|------------------------------:|----:|----:|----:|----:|----:|
| median            |  1.7 |  1.4 |  1.3 |  0.9 |  1.2 |
| 95 percentile     |  4.9 |  4.4 |  4.4 |  3.7 |  4.8 |
| 99 percentile     |   93 |   61 |   22 |   12 |   12 |

#### Node.js 24

| response time [ms]        |  XS |   S |   M |   L |  XL |
|------------------------------:|----:|----:|----:|----:|----:|
| median            |  1.5 |  1.6 |  1.3 |  1.3 |  1.8 |
| 95 percentile     |  3.7 |  4.1 |  3.4 |  3.2 |  4.1 |
| 99 percentile     |   13 |   12 |  7.2 |  5.4 |  7.7 |

#### Python 3.12

| response time [ms]        |  XS |   S |   M |   L |  XL |
|------------------------------:|----:|----:|----:|----:|----:|
| median            |  2.5 |  2.4 |  3.7 |  3.4 |  3.7 |
| 95 percentile     |   16 |  7.6 |  9.8 |  9.2 |  9.4 |
| 99 percentile     |  147 |   47 |   21 |   30 |   20 |

### Max load

#### Node.js 22

| response time [ms]        |  XS |   S |   M |   L |  XL |
|------------------------------:|----:|----:|----:|----:|----:|
| median            |  104 |   97 |   50 |   18 |   15 |
| 95 percentile     |  300 |  204 |  104 |   57 |   28 |
| 99 percentile     |  390 |  293 |  156 |   69 |   40 |

#### Node.js 24

| response time [ms]        |  XS |   S |   M |   L |  XL |
|------------------------------:|----:|----:|----:|----:|----:|
| median            |   86 |   12 |  8.2 |  6.8 |  6.3 |
| 95 percentile     |  100 |   93 |   65 |   23 |   17 |
| 99 percentile     |  191 |  102 |   73 |   31 |   27 |

#### Python 3.12

| response time [ms]        |   XS |    S |   M |   L |   XL |
|------------------------------:|-----:|-----:|----:|----:|-----:|
| median            |   902 |  699 |  384 |  137 |  119 |
| 95 percentile     |  1157 |  846 |  397 |  213 |  163 |
| 99 percentile     |  1540 | 1100 |  585 |  300 |  228 |

The bigger the runtime profile, the more resources are available to serve the response quicker. Consider these limits of the serving layer as a baseline because this does not take your Function logic into account.

### Scaling

Function runtime Pods can be scaled horizontally from zero up to the limits of the available resources at the Kubernetes worker nodes.
See the [Use External Scalers](tutorials/01-130-use-external-scalers.md) tutorial for more information.
