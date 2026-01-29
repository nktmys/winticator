install: node claude vuln go-tools

node:
    curl -fsSL https://deb.nodesource.com/setup_22.x | sudo bash -
    sudo apt-get update && sudo apt-get install -y --no-install-recommends nodejs
    sudo apt-get clean && sudo rm -rf /var/lib/apt/lists/*
    sudo npm i -g npm

claude:
    sudo chown -R $(whoami):$(whoami) ~/.claude
    curl -fsSL https://claude.ai/install.sh | bash
    curl -LsSf https://astral.sh/uv/install.sh | sh

vuln:
    curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sudo sh -s -- -b /usr/local/bin

go-tools: go-ci-tools
    go install github.com/acroca/go-symbols@latest
    go install github.com/cweill/gotests/gotests@latest
    go install github.com/fatih/gomodifytags@latest
    go install github.com/go-delve/delve/cmd/dlv@latest
    go install github.com/haya14busa/goplay/cmd/goplay@latest
    go install github.com/josharian/impl@latest
    go install github.com/nsf/gocode@latest
    go install github.com/ramya-rao-a/go-outline@latest
    go install github.com/swaggo/swag/cmd/swag@latest
    go install github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest
    go install golang.org/x/lint/golint@latest
    go install golang.org/x/tools/gopls@latest
    go install fyne.io/tools/cmd/fyne@latest

go-ci-tools:
    go install golang.org/x/vuln/cmd/govulncheck@latest
    go install honnef.co/go/tools/cmd/staticcheck@latest
