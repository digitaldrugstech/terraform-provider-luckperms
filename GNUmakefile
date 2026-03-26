default: build

build:
	go build -v ./...

install: build
	go install -v ./...

test:
	go test ./internal/... -v -count=1

testacc:
	TF_ACC=1 go test ./internal/... -v -timeout 30m -count=1

fmt:
	go fmt ./...

vet:
	go vet ./...

generate:
	go generate ./...

.PHONY: default build install test testacc fmt vet generate
