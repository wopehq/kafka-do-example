.PHONY: clean-kafka run-kafka logs-kafka
.PHONY: build-dev clean-dev run-dev logs-dev
.PHONY: clean-prod run-prod logs-prod stop-prod
.PHONY: clean-all

# KAFKA

clean-kafka:
	docker-compose -f containers/composes/dc.dev.kafka.yml down

run-kafka:
	docker-compose -f containers/composes/dc.dev.kafka.yml up -d

logs-kafka:
	docker-compose -f containers/composes/dc.dev.kafka.yml logs -f

# DEV

build-dev:
	docker build -t example -f containers/images/Dockerfile . && docker build -t example-random -f containers/images/Dockerfile.tools.random . && docker build -t example-device -f containers/images/Dockerfile.tools.device .

clean-dev:
	docker-compose -f containers/composes/dc.dev.yml down --timeout 120

run-dev:
	docker-compose -f containers/composes/dc.dev.yml up --timeout 120

logs-dev:
	docker-compose -f containers/composes/dc.dev.yml logs -f --tail 100

# PROD

clean-prod:
	docker-compose -f containers/composes/dc.prod.yml down --timeout 120

run-prod:
	docker-compose -f containers/composes/dc.prod.yml up -d --timeout 120

logs-prod:
	docker-compose -f containers/composes/dc.prod.yml logs -f --tail 100

stop-prod:
	docker-compose -f containers/composes/dc.prod.yml stop --timeout 120

# Tools

clean-all: clean-kafka clean-dev clean-prod
