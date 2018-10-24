
all: docker docker-backend

build:
	go build -o release/downscaler ./cmd/downscaler

docker: build-docker bake-docker

build-docker:
	CGO_ENABLED=0 GOOS=linux go build -o release/downscaler.linux ./cmd/downscaler

bake-docker:
	docker build -t bernardovale/downscaler:latest -f deployments/docker/downscaler/Dockerfile ./release

build-backend:
	go build -o release/default-backend ./cmd/default-backend

build-backend-docker:
	CGO_ENABLED=0 GOOS=linux go build -o release/default-backend.linux ./cmd/default-backend

bake-backend-docker:
	docker build -t bernardovale/default-backend:latest -f deployments/docker/default-backend/Dockerfile ./release

docker-backend: build-backend-docker bake-backend-docker
