default:
    go generate ../src/usecase/totpstore/migration/

install:
    sudo apt-get update && sudo apt-get install -y --no-install-recommends protobuf-compiler
    sudo apt-get clean && sudo rm -rf /var/lib/apt/lists/*
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
