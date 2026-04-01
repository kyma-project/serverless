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

Functions serve user-provided logic wrapped in the web framework, Express for Node.js and Bottle for Python. Taking the user logic aside, those frameworks have limitations and depend on the selected [runtime profile](technical-reference/07-80-available-presets.md#functions-resources) and the Kubernetes nodes specification.

The following tables present the response times of the selected runtime profiles for a "Hello World" Function across three load scenarios. This describes the overhead of the serving framework itself. Any user logic added on top of that adds extra milliseconds and must be profiled separately.

Tests are implemented using [k6](https://k6.io/) and consist of the following scenarios:

- **Constant load** — 50 virtual users send one request per second each (with a 1-second sleep between calls) for 2 minutes. Represents a steady, moderate traffic baseline.
- **Max load** — 100 virtual users send requests as fast as possible (no sleep) for 2 minutes. Represents sustained high concurrency.

> [!NOTE]
> ⚠ marks runtime profiles where errors occurred during the test scenario.

### Constant load

#### Node.js 22

| response time             | XS [ms] | S [ms] | M [ms] | L [ms] | XL [ms] |
|------------------------------:|--------:|-------:|-------:|-------:|--------:|
| median            |    2.01 |   1.97 |   1.87 |   0.90 |    0.87 |
| 95 percentile     |    4.76 |   4.39 |   5.36 |   4.04 |    3.77 |
| 99 percentile     |   94.80 |  24.80 |  18.70 |  15.90 |   10.80 |

#### Node.js 24

| response time             | XS [ms] | S [ms] | M [ms] | L [ms] | XL [ms] |
|------------------------------:|--------:|-------:|-------:|-------:|--------:|
| median            |    1.93 |   1.86 |   1.04 |   0.99 |    1.81 |
| 95 percentile     |    3.95 |   3.81 |   3.14 |   2.93 |    4.47 |
| 99 percentile     |   16.90 |   7.67 |   8.28 |   4.80 |    8.44 |

#### Python 3.12

| response time             | XS [ms] | S [ms] | M [ms] | L [ms] | XL [ms] |
|------------------------------:|--------:|-------:|-------:|-------:|--------:|
| median            |    2.61 |   2.54 |   3.73 |   3.75 |    3.69 |
| 95 percentile     |   20.40 |   7.32 |  11.60 |  10.80 |    8.66 |
| 99 percentile     |  180.00 |  46.10 |  23.80 |  21.40 |   22.20 |

### Max load

#### Node.js 22

| response time             | XS [ms] |  S [ms] |  M [ms] | L [ms] | XL [ms] |
|------------------------------:|--------:|--------:|--------:|-------:|--------:|
| median            |  102.00 |   96.70 |   49.60 |  19.90 |   14.70 |
| 95 percentile     |  299.00 |  206.00 |  104.00 |  51.60 |   27.80 |
| 99 percentile     |  403.00 |  301.00 |  170.00 |  67.90 |   40.70 |

#### Node.js 24

| response time             | XS [ms] | S [ms] | M [ms] | L [ms] | XL [ms] |
|------------------------------:|--------:|-------:|-------:|-------:|--------:|
| median            |   82.90 |  11.40 |   7.67 |   6.71 |    6.23 |
| 95 percentile     |  101.00 |  93.10 |  69.70 |  22.60 |   16.40 |
| 99 percentile     |  192.00 | 102.00 |  77.80 |  32.00 |   26.90 |

#### Python 3.12

| response time             |  XS [ms] |    S [ms] | ⚠ M [ms] |  L [ms] |  XL [ms] |
|------------------------------:|---------:|----------:|---------:|--------:|---------:|
| median            |  1000.00 |    796.00 |    88.10 |  135.00 |   124.00 |
| 95 percentile     |  1180.00 |    895.00 |   387.00 |  208.00 |   172.00 |
| 99 percentile     |  1510.00 |   1100.00 |   582.00 |  304.00 |   225.00 |

The bigger the runtime profile, the more resources are available to serve the response quicker. Consider these limits of the serving layer as a baseline because this does not take your Function logic into account.

### Scaling

Function runtime Pods can be scaled horizontally from zero up to the limits of the available resources at the Kubernetes worker nodes.
See the [Use External Scalers](tutorials/01-130-use-external-scalers.md) tutorial for more information.

