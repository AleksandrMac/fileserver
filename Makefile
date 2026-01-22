include .env.example
export

# Определяем VERSION один раз (на этапе загрузки Makefile)
MODULE_NAME := $(shell go list -m 2>/dev/null | cut -d' ' -f1)
REPO_NAME := $(shell echo "$(MODULE_NAME)" | sed 's/^[^.]*\.[^/]*\///' | tr '[:upper:]' '[:lower:]')
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
IMAGE_NAME  := cr.yandex/crp6lfi6ljf2nfcptsgo/$(REPO_NAME)

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help


help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

docker-build: ## Собрать Docker-образ
	@echo "Building $(IMAGE_NAME):$(VERSION)"
	docker build -t $(IMAGE_NAME):$(VERSION) -t $(IMAGE_NAME):latest .
.PHONY: docker-build

docker-push: docker-build ## Запушить образ в Yandex CR
	@echo "Pushing $(IMAGE_NAME):$(VERSION) and $(IMAGE_NAME):latest"
	docker push $(IMAGE_NAME):$(VERSION)
	docker push $(IMAGE_NAME):latest
.PHONY: docker-push

