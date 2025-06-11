create-queues:
	aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name event-queue --no-cli-pager

docker-up:
	docker-compose up -d

env-up: docker-up
	@echo "Waiting for LocalStack to be ready..."
	@until aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name event-queue --no-cli-pager > /dev/null 2>&1; do \
		echo "LocalStack not ready yet, waiting..."; \
		sleep 2; \
	done
	@echo "LocalStack is ready, creating queues..."
	@make create-queues

env-down:
	docker-compose down

run-synchronizer:
	go run cmd/block-synchronizer/main.go -c cmd/block-synchronizer/config.json

run-processor:
	go run cmd/event-processor/main.go -c cmd/event-processor/config.json

run-api:
	go run cmd/balance-api/main.go -c cmd/balance-api/config.json

clean-q:
	aws --endpoint-url=http://localhost:4566 sqs purge-queue --queue-url http://localhost:4566/000000000000/event-queue --no-cli-pager

