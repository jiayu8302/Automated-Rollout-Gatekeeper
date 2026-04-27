# Automated Rollout Gatekeeper (ARG)

## Overview
The **Automated Rollout Gatekeeper (ARG)** is a reliability engineering tool designed to automate the "Go/No-Go" decision process during software deployments. By integrating directly into CI/CD pipelines, ARG continuously monitors Service Level Indicators (SLIs) during a canary release and automatically halts or rolls back the deployment if anomalies are detected, effectively limiting the "blast radius" of faulty updates.

## Key Capabilities
* **Dynamic Gating**: Evaluates real-time performance metrics (Latency, Error Rate, CPU) against pre-defined safety thresholds.
* **Automated Rollback Trigger**: Dispatches signals to deployment orchestrators (like Kubernetes or Azure DevOps) to revert changes immediately upon failure.
* **Statistical Anomaly Detection**: Moves beyond static thresholds by identifying significant deviations from historical baselines.
* **Pluggable Metrics Ingestion**: Supports industry-standard telemetry providers including Prometheus, Datadog, and CloudWatch.

## Project Roadmap
- [x] **Core Scaffolding**: Initial project layout and policy definitions.
- [ ] **Phase 1**: Metric provider interface implementation for real-time data fetching.
- [ ] **Phase 2**: Sliding-window analysis for error rate spike detection.
- [ ] **Phase 3**: Integration with GitHub Actions and Spinnaker for end-to-end automation.

## Installation
```bash
go mod download
go build -o gatekeeper ./cmd/gatekeeper
