.PHONY: up clean generate build-api build-client sp_deploy test_deploy seed

help:
	@echo "HLF IoT project"
	@echo ""
	@echo "generate: generate artifacts with crypto material, configs and docker-compose templates"
	@echo "          build API and web client"
	@echo "up: bring up the network"
	@echo "clean: remove docker containers and volumes"
	@echo "build-client: build web client (building occurs inside docker container, no Node dependency)"
	@echo ""

generate: build-client build-api
	python3 network.py generate --local

up:
	python3 network.py up --local

clean:
	python3 network.py clean

build-api:
	./scripts/docker-api.sh

build-client:
	./scripts/build-client.sh

sp_deploy:
	./scripts/sp_deploy.sh

test_deploy:
	./scripts/test_deploy.sh

seed:
	./scripts/newman_seed_test.sh
