[no-cd]
default:
    CGO_ENABLED=1 go build -ldflags="-s -w" -trimpath -v -o .out/winticator ./src


# Compilation usin fyne (requires each OS)
[no-cd]
windows:
    fyne package -os windows -src=./src

[no-cd]
macos:
    fyne package -os darwin -src=./src

[no-cd]
linux:
    fyne package -os linux -src=./src


# Cross-compilation using fyne-cross (requires Docker)
[no-cd]
fyne-cross-windows:
    fyne-cross windows -arch=* -dir=./src

[no-cd]
fyne-cross-macos:
    fyne-cross darwin -arch=* -dir=./src

[no-cd]
fyne-cross-linux:
    fyne-cross linux -arch=* -dir=./src
