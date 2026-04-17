{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    # Core development tools
    go_1_26
    gcc
    pkg-config

    # HID hardware access
    hidapi
    udev
    libusb1

    # Raylib dependencies for graphics
    libGL
    libx11
    libxext
    libxcursor
    libxinerama
    libxrandr
    libxi

    # Wayland support
    wayland
    wayland-protocols
    libxkbcommon

    # Audio (optional for Raylib)
    alsa-lib

    # Development tools
    git
  ];

  shellHook = ''
    export CGO_ENABLED=1
    export GOPATH="$HOME/go"
    export PATH="$GOPATH/bin:$PATH"

    # Library paths for linking
    export PKG_CONFIG_PATH="${pkgs.hidapi}/lib/pkgconfig:${pkgs.udev}/lib/pkgconfig:${pkgs.libusb1}/lib/pkgconfig:${pkgs.wayland}/lib/pkgconfig:${pkgs.libxkbcommon}/lib/pkgconfig:$PKG_CONFIG_PATH"
    export LD_LIBRARY_PATH="${pkgs.libGL}/lib:${pkgs.wayland}/lib:${pkgs.libxkbcommon}/lib:${pkgs.udev}/lib:${pkgs.libusb1}/lib:$LD_LIBRARY_PATH"

    echo "🎹 Keyboard Visualizer (Go + Raylib)"
    echo "Tools:"
    echo "  - Go: $(go version)"
    echo "  - HID: $(pkg-config --exists hidapi-hidraw && echo 'OK' || echo 'Missing')"
    echo "  - OpenGL: $(pkg-config --exists gl && echo 'OK' || echo 'Missing')"
    echo "  - Wayland: $(pkg-config --exists wayland-client && echo 'OK' || echo 'Missing')"
    echo ""
    echo "Build: cd app-raylib && go build -o keyboard-visualizer ."
  '';
}