BINARY_NAME=integritySum

## Runs all of the required cleaning and verification targets.
.PHONY : all
all: mod-download test dev-dependencies

## Downloads the Go module.
.PHONY : mod-download
mod-download:
	@echo "==> Downloading Go module"
	go mod download

## Runs the test suite with mocks enabled.
.PHONY : test
test:
	go test -v internal/core/services/hash_test.go

## Downloads the necessesary dev dependencies.
.PHONY : dev-dependencies
dev-dependencies: minikube docker helm
	@echo "==> Downloaded development dependencies"
	@echo "==> Successfully installed"

.PHONY : minikube
minikube:
	minikube start

.PHONY : docker
docker:
	@eval $$(minikube docker-env) ;\
    docker image build -t hasher:latest -f Dockerfile .

.PHONY : helm
helm:
	helm dependency update helm-charts/
	helm install app helm-charts/

.PHONY : tidy
tidy: ## Cleans the Go module.
	@echo "==> Tidying module"
	@go mod tidy

.PHONY : build
build:
	go build -o ${BINARY_NAME} cmd/main.go

.PHONY : run
run:
	go build -o ${BINARY_NAME} cmd/main.go
	./${BINARY_NAME}

## Remove an installed Helm deployment and stop minikube.
stop:
	helm uninstall app
	minikube stop

## Cleaning build cache.
.PHONY : clean
clean:
	go clean
	rm ${BINARY_NAME}

## Compile the binary.
compile:
	echo "Building for Windows 32-bit"
	GOOS=windows GOARCH=386 go build -o bin/${BINARY_NAME}_32bit.exe cmd/main.go

	echo "Building for Windows 64-bit"
	GOOS=windows GOARCH=amd64 go build -o bin/${BINARY_NAME}_64bit.exe cmd/main.go

	echo "Building for Linux 32-bit"
	GOOS=linux GOARCH=386 go build -o bin/${BINARY_NAME}_32bit cmd/main.go

	echo "Building for Linux 64-bit"
	GOOS=linux GOARCH=amd64 go build -o bin/${BINARY_NAME} cmd/main.go

	echo "Building for MacOS X 64-bit"
	GOOS=darwin GOARCH=amd64 go build -o bin/${BINARY_NAME}_macos cmd/main.go
