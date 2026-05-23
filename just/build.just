set working-directory := '..'

default:
    CGO_ENABLED=1 go build -ldflags="-s -w" -trimpath -v -o .out/winticator ./src


# Compilation usin fyne (requires each OS)
windows:
    fyne package -os windows -src=./src

macos:
    fyne package -os darwin -src=./src

linux:
    fyne package -os linux -src=./src


# Cross-compilation using fyne-cross (requires Docker)
fyne-cross-windows:
    fyne-cross windows -arch=* -dir=./src

fyne-cross-macos:
    fyne-cross darwin -arch=* -dir=./src

fyne-cross-linux:
    fyne-cross linux -arch=* -dir=./src
