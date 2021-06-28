.DEFAULT_GOAL := help

.PHONY: help
help: ## show this help
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[33m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: sync
sync: ## start sync unidirectionally with dev server
	rsync -chavzP --stats . gordonpn@dev:workspace/cloudflare-ddns;
	fswatch -o . | while read f; do rsync -chavzP --stats . gordonpn@dev:workspace/cloudflare-ddns; done

.PHONY: build-docker
build-docker: ## build docker image for arm64 and amd64
	docker buildx build --platform linux/amd64,linux/arm64 .

.PHONY: run-docker
run-docker: ## build docker image and run
	DOCKER_BUILDKIT=1 docker build -t cloudflare-ddns . && docker run -it --rm cloudflare-ddns

.PHONY: lint
lint: ## lint using golangci-lint
	golangci-lint run
