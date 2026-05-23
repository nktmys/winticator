set working-directory := '..'

install: node claude vuln go-tools

node:
    curl -fsSL https://deb.nodesource.com/setup_24.x | sudo bash -
    sudo apt-get update && sudo apt-get install -y --no-install-recommends nodejs
    sudo apt-get clean && sudo rm -rf /var/lib/apt/lists/*

claude:
    curl -fsSL https://claude.ai/install.sh | bash
    curl -LsSf https://astral.sh/uv/install.sh | sh

vuln:
    go install golang.org/x/vuln/cmd/govulncheck@v1.3.0
    curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sudo sh -s -- -b /usr/local/bin v0.70.0

go-tools:
    go install github.com/abice/go-enum@v0.9.2
    go install github.com/cweill/gotests/gotests@v1.9.0
    go install github.com/fatih/gomodifytags@v1.17.0
    go install github.com/go-delve/delve/cmd/dlv@v1.26.3
    go install github.com/josharian/impl@v1.5.0
    go install golang.org/x/tools/gopls@v0.22.0
    go install fyne.io/tools/cmd/fyne@v1.7.1
    curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b ${GOBIN:-$(go env GOPATH)/bin} v2.12.2
