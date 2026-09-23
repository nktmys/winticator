set working-directory := '..'

default:
    git config --global --add safe.directory /app
    xdg-user-dirs-update
