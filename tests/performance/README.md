# Performance Tests

Performance tests measure the response time overhead of the Serverless serving layer across all runtime profiles using a "Hello World" Function.

## Prerequisites

- Access to a Kubernetes cluster with Serverless installed
- `kubectl` configured to point to the target cluster

## Setup

1. Install the monitoring stack in the cluster:

   ```bash
   make install-monitoring
   ```

2. Forward the Grafana port to localhost:

   ```bash
   make forward-grafana
   ```

## Running the Tests

1. Run the full test suite three times to collect enough data for averaging:

   ```bash
   make start-test
   ```

2. Follow the progress in a separate terminal:

   ```bash
   make follow-remote-test
   ```

Each run tests all runtime/profile combinations sequentially.

## Collecting Results

After each run, open Grafana at `http://localhost:3000` and export the results:

1. Open the **Serverless Performance Tests** dashboard.
2. Set the time range to cover exactly one completed test run — from just before the first scenario started to just after the last one finished. The time range must not overlap with other runs, as the **Test Results Summary** panel aggregates data using `min`/`max` over the selected range and will produce incorrect results if multiple runs are included.
3. Open the **Test Results Summary** panel.
4. Use the panel menu → **Inspect** → **Data** → **Download CSV**. Make sure the **Apply panel transformations** toggle is enabled — without it, the exported data will be raw time series instead of the aggregated summary table.

## Updating the Documentation

After collecting CSVs from three runs, update `docs/user/00-50-limitations.md`.

You can use an AI assistant such as Claude Code to automate this step. Example prompt:

```
claude "Update the performance test results in docs/user/00-50-limitations.md based on the CSV files: <path1.csv>, <path2.csv>, <path3.csv>. Recalculate averages across all runs (exclude outliers if a single value deviates more than 3x from the others). If any runtime/profile combination shows a significant error rate, note it in the document — use your judgement on what constitutes a significant error rate. Also update the node specification in the NOTE if it differs from the current cluster — check it with: kubectl get nodes -o wide and kubectl get nodes -o jsonpath='{.items[0].status.nodeInfo.kernelVersion}'. Update the test run count and last updated date accordingly."
```

Replace `<path1.csv>`, `<path2.csv>`, `<path3.csv>` with the actual paths to your downloaded CSV files.
