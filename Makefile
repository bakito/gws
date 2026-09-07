# Include toolbox tasks
include ./.toolbox.mk

# Run go golanci-lint
lint: tb.golangci-lint
	$(TB_GOLANGCI_LINT) run --fix

# Run go mod tidy
tidy:
	go mod tidy

# Run tests
test:
	go test ./... -v -coverprofile=coverage.out

release: tb.goreleaser tb.semver
	@version=$$($(TB_SEMVER)); \
	git tag -s $$version -m"Release $$version"; \
	git push origin $$version
	$(TB_GORELEASER) --clean

test-release: tb.goreleaser
	$(TB_GORELEASER) --skip=publish --snapshot --clean

fmt: tb.golines tb.gofumpt
	$(TB_GOLINES) --base-formatter="$(TB_GOFUMPT)" --max-len=120 --write-output .

build-win:
	GOOS=windows GOARCH=amd64 go build -o gws.exe -ldflags="-s -w -X github.com/bakito/gws/version.Version=dev-$$(date +%Y%m%d-%H%M)" .

check-vulnerabilities:
	go run golang.org/x/vuln/cmd/govulncheck@latest -show verbose,color ./...