GIT_VERSION ?= $(shell git describe --abbrev=8 --tags --always --dirty 2>/dev/null || echo "dev")
IMAGE_PREFIX ?= bladedancer
SERVICE_NAME  = extauthzdemo

.PHONY: default
default: local.build

.PHONY: clean
clean:
	go clean
	rm -f bin/$(SERVICE_NAME)

.PHONY: local.build
local.build: clean
	GOARCH=amd64 GOOS=linux go build -o bin/$(SERVICE_NAME) main.go

.PHONY: local.test
local.test:
	go test -v ./...

.PHONY: local.run
local.run:
	go run ./main.go --port 10001

.PHONY: docker.build
docker.build:
	docker build -t $(IMAGE_PREFIX)/$(SERVICE_NAME) -f ./Dockerfile .
	docker tag $(IMAGE_PREFIX)/$(SERVICE_NAME) $(IMAGE_PREFIX)/$(SERVICE_NAME):$(GIT_VERSION)

.PHONY: docker.run
docker.run:
	docker run -p 10001:10001 $(IMAGE_PREFIX)/$(SERVICE_NAME)

.PHONY: docker.compose.up
docker.compose.up:
	docker compose up --build

.PHONY: dep
dep:
	go mod tidy

.PHONY: vet
vet:
	go vet ./...
