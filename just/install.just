install: node claude vuln go-tools

node:
    curl -fsSL https://deb.nodesource.com/setup_24.x | sudo bash -
    sudo apt-get update && sudo apt-get install -y --no-install-recommends nodejs
    sudo apt-get clean && sudo rm -rf /var/lib/apt/lists/*
    sudo npm i -g npm

claude:
    git config --global --add safe.directory /app
    sudo chown -R $(whoami):$(whoami) ~/.claude
    curl -fsSL https://claude.ai/install.sh | bash
    curl -LsSf https://astral.sh/uv/install.sh | sh

vuln:
    go install golang.org/x/vuln/cmd/govulncheck@latest
    curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sudo sh -s -- -b /usr/local/bin

go-tools:
    go install github.com/abice/go-enum@latest
    go install github.com/cweill/gotests/gotests@latest
    go install github.com/fatih/gomodifytags@latest
    go install github.com/go-delve/delve/cmd/dlv@latest
    go install github.com/josharian/impl@latest
    go install golang.org/x/tools/gopls@latest
    go install fyne.io/tools/cmd/fyne@latest
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin
