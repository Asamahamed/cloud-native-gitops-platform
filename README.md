# ☁️ Cloud-Native GitOps Platform with Full Observability

A production-style, end-to-end DevOps/SRE platform on **AWS EKS**, provisioned with
**Terraform**, delivered via **GitOps (ArgoCD)**, secured with a **DevSecOps CI/CD
pipeline**, and monitored with a complete **observability stack** (Prometheus,
Grafana, Loki, Alertmanager).

> Built to demonstrate real-world platform engineering: infrastructure-as-code,
> continuous delivery, zero-downtime rollouts, autoscaling, and SLO-based alerting.

---

## 🏛️ Architecture

```mermaid
flowchart LR
    dev[Developer] -->|git push| gh[GitHub Repo]

    subgraph CI/CD [GitHub Actions - DevSecOps]
        gh --> test[Go build + tests]
        test --> sec[Trivy + Checkov scans]
        sec --> img[Build image -> GHCR]
        img --> bump[Bump image tag in Git]
    end

    bump -->|Git is source of truth| argo[ArgoCD]

    subgraph EKS [AWS EKS Cluster - Terraform]
        argo -->|sync| staging[demo-staging ns]
        argo -->|sync| prod[demo-production ns]
        argo -->|sync| mon[monitoring ns]

        staging --> app[demo-app pods + HPA]
        app -->|/metrics| prom[Prometheus]
        app -->|logs| loki[Loki]
        prom --> graf[Grafana]
        loki --> graf
        prom --> alert[Alertmanager]
    end
```

**Flow:** push code → CI runs tests + security scans → image built, scanned and
pushed to GHCR → CI bumps the image tag in Git → **ArgoCD** detects the change and
syncs the cluster → Prometheus/Loki observe the app → Grafana visualises the
golden signals → Alertmanager fires on SLO breaches.

---

## 🧰 Tech Stack

| Layer | Tools |
|-------|-------|
| **Infrastructure** | Terraform, AWS (EKS, VPC, IAM, NAT, EC2) |
| **Containers** | Docker (multi-stage, distroless, non-root) |
| **Orchestration** | Kubernetes, Kustomize (base + overlays), HPA, PDB |
| **GitOps / CD** | ArgoCD (app-of-apps pattern) |
| **CI / DevSecOps** | GitHub Actions, Trivy, Checkov |
| **Observability** | Prometheus, Grafana, Loki, Alertmanager, ServiceMonitor |
| **App** | Go (Prometheus metrics, health/readiness probes, graceful shutdown) |

---

## 📁 Repository Layout

```
cloud-native-gitops-platform/
├── app/                     # Go microservice + Dockerfile + tests
├── terraform/               # AWS VPC + EKS infrastructure as code
├── kubernetes/
│   ├── base/                # Deployment, Service, HPA, PDB, ServiceMonitor
│   └── overlays/            # staging + production (Kustomize)
├── argocd/
│   ├── app-of-apps.yaml     # single root app that manages everything
│   └── apps/                # Prometheus stack, Loki, demo app (staging/prod)
├── observability/           # SLO alert rules + Grafana dashboard
├── .github/workflows/       # CI/CD pipeline
└── Makefile                 # one-command bootstrap
```

---

## 🚀 Quick Start

### Prerequisites
`aws-cli` · `terraform >= 1.6` · `kubectl` · `docker` · `go >= 1.23`

---

### 🟢 Run & test locally (free — no cloud needed)

Everything below runs on your laptop and proves the code is correct without
spending a cent on AWS.

**1. Test & run the Go service**
```bash
cd app
go test ./...            # run unit tests  → "ok ..."
go run .                 # start the service → http://localhost:8080
```
> **Port 8080 already in use?** Pick another port:
> - macOS/Linux: `PORT=8090 go run .`
> - Windows (cmd): `set PORT=8090 && go run .`
> - Windows (PowerShell): `$env:PORT=8090; go run .`

**2. See it running** — open in a browser or curl:
```bash
curl http://localhost:8090/            # service info + version
curl http://localhost:8090/healthz     # {"status":"ok"}   (liveness)
curl http://localhost:8090/readyz      # {"status":"ready"} (readiness)
curl http://localhost:8090/metrics     # Prometheus metrics
```
> This is a **backend API** — a JSON response (not an error) means it works. The
> visual Grafana/ArgoCD dashboards only appear once deployed to a cluster (below).

**3. Validate the infrastructure (no cluster required)**
```bash
docker build -t gitops-demo-app app                # build the container image
kubectl kustomize kubernetes/overlays/staging      # render the K8s manifests
terraform -chdir=terraform init -backend=false     # download modules
terraform -chdir=terraform validate                # validate the IaC
terraform -chdir=terraform fmt -check              # check formatting
```

---

### 🔴 Deploy for real on AWS  ⚠️ *creates billable resources*

### 2. Provision the cluster
```bash
make infra-init
make infra-apply        # creates VPC + EKS (takes ~15 min)
make kubeconfig
```

### 3. Bootstrap GitOps
```bash
make bootstrap          # installs ArgoCD + applies the app-of-apps
```
ArgoCD now reconciles the **entire platform** — monitoring stack + app — straight
from this Git repo. Nothing is deployed by hand.

### 4. Explore
```bash
kubectl -n argocd port-forward svc/argocd-server 8080:443      # ArgoCD UI
kubectl -n monitoring port-forward svc/kube-prometheus-stack-grafana 3000:80  # Grafana
```

### 5. Tear down
```bash
make infra-destroy
```

---

## 🔐 Reliability & Security Highlights
- **Zero-downtime rollouts** — `maxUnavailable: 0`, readiness gates, PodDisruptionBudget
- **Autoscaling** — HPA on CPU & memory (2 → 10 replicas)
- **Hardened containers** — distroless, non-root, read-only root FS, all caps dropped
- **Shift-left security** — Trivy (images + filesystem) and Checkov (IaC) gate every PR
- **SLO alerting** — error-rate, p95 latency and availability alerts in Prometheus
- **Least-privilege** — IRSA enabled for pod-level IAM

---

## 📊 Golden Signals Monitored
Latency (p50/p95/p99) · Traffic (req/s) · Errors (5xx ratio) · Saturation (replicas / HPA)

---

## 👤 Author
**Asam Ahamed** — DevOps & Site Reliability Engineer
[GitHub](https://github.com/Asamahamed) · asamahamed487@gmail.com
