default: lint test
all: lint test vuln
    just build

mod init 'just/init.just'
mod install 'just/install.just'
mod build 'just/build.just'
mod protobuf 'just/protobuf.just'

lint:
    golangci-lint fmt ./...
    golangci-lint run ./...

test:
    go test ./... -failfast

vuln:
    GOMEMLIMIT=256MiB govulncheck ./...
    trivy fs --quiet --config .trivy/trivy.yaml ./

setup:
    just init
    just install
