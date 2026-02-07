default: lint test
all: lint test vuln
    just build

mod install 'just/install.just'
mod build 'just/build.just'
mod protobuf 'just/protobuf.just'

lint:
    go fmt ./...
    go mod tidy
    staticcheck ./...

test:
    go test ./... -failfast

vuln:
    GOMEMLIMIT=256MiB govulncheck ./...
    trivy fs --quiet --config trivy/trivy.yaml ./
