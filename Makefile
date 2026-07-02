.PHONY: help test run docker infra-init infra-apply infra-destroy kubeconfig argocd-install bootstrap

CLUSTER ?= gitops-platform
REGION  ?= us-east-1

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

test: ## Run the Go unit tests
	cd app && go test -race ./...

run: ## Run the service locally on :8080
	cd app && go run .

docker: ## Build the container image locally
	docker build -t gitops-demo-app:local app

infra-init: ## terraform init
	cd terraform && terraform init

infra-apply: ## Provision the EKS cluster
	cd terraform && terraform apply -auto-approve

infra-destroy: ## Tear down all infrastructure
	cd terraform && terraform destroy -auto-approve

kubeconfig: ## Point kubectl at the new cluster
	aws eks update-kubeconfig --region $(REGION) --name $(CLUSTER)

argocd-install: ## Install ArgoCD into the cluster
	kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

bootstrap: argocd-install ## Install ArgoCD and apply the app-of-apps
	kubectl apply -f argocd/app-of-apps.yaml
	@echo "✅ Platform bootstrapped. ArgoCD will reconcile everything from Git."
