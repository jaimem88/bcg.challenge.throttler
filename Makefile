# Project Name
SHA1				:= $(shell git rev-parse --verify --short HEAD)
INTERNAL_BUILD_ID	:= $(shell [ -z "${BUILD_ID}" ] && echo "local" || echo ${BUILD_ID})
VERSION				:= $(shell echo "${INTERNAL_BUILD_ID}_${SHA1}")
BINARY				:= $(shell basename -s .git `git config --get remote.origin.url`)

CGO_ENABLED :=0
GOOS:=linux
.PHONY: config
config:
	$(shell go build ./cmd/$(BINARY) && ./$(BINARY) -default config.json)

.PHONY: build
build: config
	docker build -t jaimemartinez88/throttler:$(VERSION) .

.PHONY: run-docker
run-docker: build
	docker rm -f bcg-challenge-throttler || true
	docker run --name bcg-challenge-throttler -p $(PORT):$(PORT) -e PORT=$(PORT) throttler:$(VERSION)
ifndef PORT
	$(error PORT environment variable is missing)
endif