PACKAGES := \
	github.com/golang/dep/cmd/dep
REGISTRY_PATH       = hub.docker.com/r/ame421/nburunova
REGISTRY_NAMESPACE  = navi
APP_IMAGE_NAME      = taxa-app
REF_NAME           ?= $(shell git rev-parse --abbrev-ref HEAD)
IMAGE_VERSION      ?= ${REF_NAME}-$(shell git rev-parse HEAD)
APP_IMAGE_PATH      = ${REGISTRY_PATH}/${REGISTRY_NAMESPACE}/${APP_IMAGE_NAME}

export IMAGE_VERSION

SHELL := env PATH=$(PATH) /bin/bash

PROJECT_PKGS := $$(go list ./...)

.PHONY: install-packages
install-packages:
	$(foreach pkg,$(PACKAGES),go get -u $(pkg);)

.PHONY: dependencies
dependencies:
	dep ensure -vendor-only

.PHONY: clean-api
clean-api:
	rm -rf ./src/cmd/api/bin/*

.PHONY: build-api
build-api: clean-api
	CGO_ENABLED=0 GOOS=linux go build -o ./src/cmd/api/bin/taxa ./src/cmd/api

.PHONY: build-app
build-app:
	docker-compose run --rm taxa-build

.PHONY: build-app-image
build-app-image:
	docker-compose build taxa

.PHONY: push-app-image
push-app-image:
	docker-compose push taxa

.PHONY: tag-app-image
tag-app-image:
	docker tag $(APP_IMAGE_PATH):$(IMAGE_VERSION) $(APP_IMAGE_PATH):$(TAG)

.PHONY: registry-cleanup
registry-cleanup:
	hub-tool tags:cleanup --path ${APP_IMAGE_PATH} --regexp '^master-.*' --count 10
	hub-tool tags:cleanup --path ${APP_IMAGE_PATH} --regexp '^(?!master).*' --days 7

.PHONY: run-api
run-api:
	src/cmd/api/bin/taxa --settings src/cmd/api/settings.json

.PHONY: test
test:
	for pkg in $(PROJECT_PKGS); do \
		go test -v -race $$pkg || exit 1 ; \
	done;

.PHONY: run-image-api-mocked
run-image-api-mocked:
	docker run --net=host --rm --name=taksa  --entrypoint "/usr/lib/nbu421/taksa" $(API_IMAGE_NAME) --settings /etc/nbu421/settings.json --mock-enabled

.PHONY: lint
lint:
	gometalinter --config=gometalinter.json ./...
