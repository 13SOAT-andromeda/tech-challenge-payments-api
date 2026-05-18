CLUSTER_NAME := payments
IMAGE        := payments-api:latest
OVERLAY      := k8s/overlays/local

.PHONY: help cluster-up cluster-down ingress build load deploy up restart logs status down destroy

help: ## Mostra esta ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# ── Cluster ───────────────────────────────────────────────────────────────────

cluster-up: ## Cria o cluster kind com port mappings para ingress
	kind create cluster --name $(CLUSTER_NAME) --config kind-config.yaml

cluster-down: ## Remove o cluster kind
	kind delete cluster --name $(CLUSTER_NAME)

ingress: ## Instala ingress-nginx e aguarda o controller ficar pronto
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
	kubectl rollout status deployment/ingress-nginx-controller \
		--namespace ingress-nginx \
		--timeout=90s

# ── Imagem ────────────────────────────────────────────────────────────────────

build: ## Build da imagem Docker
	docker build -t $(IMAGE) .

load: ## Carrega a imagem no cluster kind
	kind load docker-image $(IMAGE) --name $(CLUSTER_NAME)

# ── Deploy ────────────────────────────────────────────────────────────────────

deploy: ## Aplica o overlay local no cluster
	kubectl apply -k $(OVERLAY)

restart: ## Reinicia o deployment payments-api (útil após rebuild)
	kubectl rollout restart deployment/payments-api
	kubectl rollout status deployment/payments-api --timeout=60s

down: ## Remove todos os recursos do overlay local
	kubectl delete -k $(OVERLAY) --ignore-not-found

# ── Observabilidade ───────────────────────────────────────────────────────────

status: ## Mostra pods, services e ingress
	@echo "\n=== Pods ==="
	kubectl get pods
	@echo "\n=== Services ==="
	kubectl get svc
	@echo "\n=== Ingress ==="
	kubectl get ingress

logs: ## Acompanha os logs do payments-api
	kubectl logs -f -l app=payments-api --tail=50

# ── Workflows compostos ───────────────────────────────────────────────────────

up: cluster-up ingress build load deploy ## Setup completo: cluster + ingress + build + load + deploy
	@echo "\n✓ Cluster pronto. Aguardando postgres..."
	kubectl wait --for=condition=ready pod --selector=app=postgres --timeout=60s
	@echo "✓ API disponível em http://localhost/health"

rebuild: build load restart ## Rebuild e redeploy sem recriar o cluster

destroy: down cluster-down ## Remove recursos e deleta o cluster
