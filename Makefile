CLUSTER_NAME := payments
IMAGE        := payments-api:latest
OVERLAY      := k8s/overlays/local
PORT         := 8080

SHELL := /bin/bash

.PHONY: help cluster-up cluster-down ingress build load deploy up restart logs status down destroy port-forward localstack-forward localstack-setup localstack-endpoints localstack-apigw

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

port-forward: ## Expõe a API em localhost:$(PORT) via port-forward (necessário no WSL2)
	kubectl port-forward svc/payments-api-svc $(PORT):80

localstack-forward: ## Expõe o LocalStack em localhost:4566 via port-forward
	kubectl port-forward svc/localstack 4566:4566

localstack-endpoints: ## Cria Endpoints no cluster apontando para o LocalStack externo (docker-compose)
	@HOST_GATEWAY=$$(docker network inspect kind --format '{{range .IPAM.Config}}{{.Gateway}}{{end}}' 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' | head -1); \
	if [ -z "$$HOST_GATEWAY" ]; then \
		echo "Erro: rede kind não encontrada. Execute make cluster-up primeiro." && exit 1; \
	fi; \
	echo "Apontando localstack service para $$HOST_GATEWAY:4566"; \
	HOST_GATEWAY=$$HOST_GATEWAY envsubst < $(OVERLAY)/localstack-endpoints.yaml.tpl | kubectl apply -f -

localstack-apigw: ## Cria REST API Gateway no LocalStack com rota pública POST /webhooks/mercadopago
	@ENDPOINT=http://localhost:4566; \
	REGION=us-east-1; \
	API_NAME=payments-webhook-api; \
	EXISTING_ID=$$(aws --endpoint-url=$$ENDPOINT --region $$REGION \
		apigateway get-rest-apis \
		--query "items[?name=='$$API_NAME'].id" --output text 2>/dev/null); \
	if [ -n "$$EXISTING_ID" ]; then \
		echo "API Gateway '$$API_NAME' já existe ($$EXISTING_ID). Removendo para recriar..."; \
		aws --endpoint-url=$$ENDPOINT --region $$REGION apigateway delete-rest-api --rest-api-id $$EXISTING_ID; \
	fi; \
	echo "Criando REST API Gateway: $$API_NAME"; \
	API_ID=$$(aws --endpoint-url=$$ENDPOINT --region $$REGION \
		apigateway create-rest-api \
		--name $$API_NAME \
		--query 'id' --output text); \
	ROOT_ID=$$(aws --endpoint-url=$$ENDPOINT --region $$REGION \
		apigateway get-resources --rest-api-id $$API_ID \
		--query 'items[?path==`/`].id' --output text); \
	WH_ID=$$(aws --endpoint-url=$$ENDPOINT --region $$REGION \
		apigateway create-resource \
		--rest-api-id $$API_ID --parent-id $$ROOT_ID --path-part webhooks \
		--query 'id' --output text); \
	MP_ID=$$(aws --endpoint-url=$$ENDPOINT --region $$REGION \
		apigateway create-resource \
		--rest-api-id $$API_ID --parent-id $$WH_ID --path-part mercadopago \
		--query 'id' --output text); \
	aws --endpoint-url=$$ENDPOINT --region $$REGION \
		apigateway put-method \
		--rest-api-id $$API_ID --resource-id $$MP_ID \
		--http-method POST --authorization-type NONE > /dev/null; \
	echo "Criando integração → http://host.docker.internal:80/webhooks/mercadopago"; \
	aws --endpoint-url=$$ENDPOINT --region $$REGION \
		apigateway put-integration \
		--rest-api-id $$API_ID --resource-id $$MP_ID \
		--http-method POST --type HTTP_PROXY \
		--integration-http-method POST \
		--uri http://host.docker.internal:80/webhooks/mercadopago \
		--passthrough-behavior WHEN_NO_MATCH > /dev/null; \
	aws --endpoint-url=$$ENDPOINT --region $$REGION \
		apigateway create-deployment \
		--rest-api-id $$API_ID --stage-name local > /dev/null; \
	echo ""; \
	echo "✓ API Gateway pronto. Somente POST /webhooks/mercadopago está exposto."; \
	echo "  Webhook público: http://$$API_ID.execute-api.localhost.localstack.cloud:4566/local/webhooks/mercadopago"

localstack-setup: ## Cria fila SQS e tópico SNS no LocalStack local (docker-compose)
	@echo "Criando fila SQS: payment-order-events-queue"
	aws --endpoint-url=http://localhost:4566 --region us-east-1 \
		sqs create-queue --queue-name payment-order-events-queue
	@echo "Criando tópico SNS: payment-events"
	aws --endpoint-url=http://localhost:4566 --region us-east-1 \
		sns create-topic --name payment-events
	@echo "Feito. Recursos disponíveis no LocalStack."

# ── Workflows compostos ───────────────────────────────────────────────────────

up: cluster-up ingress build load deploy ## Setup completo: cluster + ingress + build + load + deploy
	@echo "\n✓ Cluster pronto. Aguardando postgres..."
	kubectl wait --for=condition=ready pod --selector=app=postgres --timeout=120s
	@echo "✓ Aguardando LocalStack..."
	kubectl wait --for=condition=ready pod --selector=app=localstack --timeout=120s
	@echo "✓ Aguardando LocalStack init (cria fila e tópico)..."
	kubectl wait --for=condition=complete job/localstack-init --timeout=120s
	@echo "✓ Aguardando payments-api..."
	kubectl rollout status deployment/payments-api --timeout=120s
	@echo "\n✓ Cluster pronto!"
	@echo "  Ingress:      http://localhost/health"
	@echo "  Port-forward: make port-forward  →  http://localhost:$(PORT)/health  (WSL2)"

rebuild: build load restart ## Rebuild e redeploy sem recriar o cluster

destroy: down cluster-down ## Remove recursos e deleta o cluster
