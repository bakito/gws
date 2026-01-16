# Include toolbox tasks
include ./.toolbox.mk

# Run go golanci-lint
lint: tb.golangci-lint
	$(TB_GOLANGCI_LINT) run --fix

# Run go mod tidy
tidy:
	go mod tidy

# Run tests
test: tb.ginkgo
	$(TB_GINKGO) -r --cover --coverprofile=coverage.out

release: tb.goreleaser tb.semver
	@version=$$($(TB_SEMVER)); \
	git tag -s $$version -m"Release $$version"
	$(TB_GORELEASER) --clean

test-release: tb.goreleaser
	$(TB_GORELEASER) --skip=publish --snapshot --clean

fmt: tb.golines tb.gofumpt
	$(TB_GOLINES) --base-formatter="$(TB_GOFUMPT)" --max-len=120 --write-output .

build-win:
	GOOS=windows GOARCH=amd64 go build -o gws.exe -ldflags="-s -w -X github.com/bakito/gws/version.Version=dev-$$(date +%Y%m%d-%H%M)" .

extract-oauth-vars:
	docker build -t auth_config.go --no-cache py
	docker create --name auth_config auth_config.go true
	docker cp auth_config:/auth_config.go internal/gcloud/auth_config.go
	docker rm auth_config
	docker rmi auth_config.go

dummy-oauth-vars:
	cd py && 	go run main.go > ../internal/gcloud/auth_config.go
