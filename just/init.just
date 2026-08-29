set working-directory := '..'

default:
    sudo chown -R $(whoami):$(whoami) ~/.config
    sudo chown -R $(whoami):$(whoami) ~/.claude
    git config --global --add safe.directory /app
    xdg-user-dirs-update
