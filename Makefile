# ShopHub backend — pomoćne komande (FZ 2)

.PHONY: deps vet test test-integration cover swagger migrate-up migrate-down docker

deps:
	go mod tidy

vet:
	go vet ./...

# Unit testovi (bez Docker-a / klastera)
test:
	go test ./... -count=1

# Integracioni testovi (Testcontainers + envtest). Zahteva Docker i KUBEBUILDER_ASSETS.
test-integration:
	go test -tags=integration ./... -count=1 -v

# Coverage (DoD: >= 60%)
cover:
	go test -tags=integration -coverpkg=./internal/...,./pkg/... -coverprofile=cover.out ./...
	go tool cover -func=cover.out | tail -1

# Swagger (swag generiše OpenAPI iz Go @-anotacija u docs/)
swagger:
	go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g cmd/shophub/main.go -o docs --parseInternal

# golang-migrate CLI (alternativa ugrađenom runner-u na startu aplikacije)
DB_URL ?= postgres://shophub:shophub@localhost:5432/shophub?sslmode=disable
migrate-up:
	migrate -path migrations -database "$(DB_URL)" up
migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

docker:
	docker build --build-arg VERSION=v0.1.0 -t ghcr.io/shophub-platform/shophub:v0.1.0 .
