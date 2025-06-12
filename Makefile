.PHONY: help env-up env-down docker-up create-queues run-synchronizer run-processor run-api clean-q enter-db

help: ## Show this help message
	@echo "Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

create-queues: ## Create SQS queues (delete existing ones first)
	-aws --endpoint-url=http://localhost:4566 sqs delete-queue --queue-url http://localhost:4566/000000000000/event-queue --no-cli-pager
	-aws --endpoint-url=http://localhost:4566 sqs delete-queue --queue-url http://localhost:4566/000000000000/test-queue --no-cli-pager
	aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name event-queue --attributes VisibilityTimeout=3 --no-cli-pager
	aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name test-queue --attributes VisibilityTimeout=3 --no-cli-pager

docker-up: ## Start Docker Compose services
	docker-compose up -d

env-up: docker-up ## Start all infrastructure services (PostgreSQL, SQS(LocalStack), Redis)
	@echo "Waiting for LocalStack to be ready..."
	@until aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name event-queue --no-cli-pager > /dev/null 2>&1; do \
		echo "LocalStack not ready yet, waiting..."; \
		sleep 2; \
	done
	@echo "LocalStack is ready, creating queues..."
	@make create-queues

env-down:
	docker-compose down

run-synchronizer: ## Run block synchronizer service (logs to ./logs/synchronizer.log)
	go run cmd/block-synchronizer/main.go -c cmd/block-synchronizer/config.json > ./logs/synchronizer.log

run-processor: ## Run event processor service (logs to ./logs/processor.log)
	go run cmd/event-processor/main.go -c cmd/event-processor/config.json > ./logs/processor.log

run-api: ## Run balance API service (logs to ./logs/api.log)
	go run cmd/balance-api/main.go -c cmd/balance-api/config.json > ./logs/api.log

clean-q: ## Clear all messages from the event queue
	aws --endpoint-url=http://localhost:4566 sqs purge-queue --queue-url http://localhost:4566/000000000000/event-queue --no-cli-pager

enter-db: ## Connect to PostgreSQL database
	psql -h localhost -p 5432 -U postgres -d onbloc