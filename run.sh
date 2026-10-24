#!/usr/bin/env sh
set -eu
umask 077

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$SCRIPT_DIR"

MODE="desktop"
REBUILD="0"
BUILD_ONLY="0"
NO_BUILD="0"

usage() {
    cat <<'EOF'
Ratel Ledger run script

Usage:
    ./run.sh [options]

Options:
    -r, --rebuild       Rebuild the frontend (and backend) before launching
        --no-build      Skip all build steps
    -b, --build-only    Build backend/frontend and exit
    -s, --server        Run only the Go backend server
    -d, --desktop       Launch the Electron desktop app (default)
    -h, --help          Show this help

Examples:
    ./run.sh                  Build if needed, then launch the desktop app
    ./run.sh --rebuild        Force a full frontend/backend rebuild
    ./run.sh --server         Run the backend on 127.0.0.1:8080
    ./run.sh -b               Build artifacts without launching
EOF
}

while [ $# -gt 0 ]; do
    case "$1" in
        --rebuild | -r)
            REBUILD="1"
            ;;
        --no-build)
            NO_BUILD="1"
            ;;
        --build-only | -b)
            BUILD_ONLY="1"
            ;;
        --server | -s)
            MODE="server"
            ;;
        --desktop | -d)
            MODE="desktop"
            ;;
        --help | -h)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            usage >&2
            exit 2
            ;;
    esac

    shift 1
done

require_command() {
    if ! command -v "$1" > /dev/null 2>&1; then
        echo "Error: \"$1\" is required but was not found in PATH." >&2
        exit 127
    fi
}

mkdir -p data log storage

if [ "$NO_BUILD" = "0" ]; then
    require_command go

    echo "==> Building backend (ratel-ledger)"
    go build -o ratel-ledger .
fi

if [ ! -x ratel-ledger ]; then
    echo "Error: backend binary \"ratel-ledger\" is missing. Run without --no-build." >&2
    exit 1
fi

if [ "$MODE" = "desktop" ] && [ "$NO_BUILD" = "0" ] && { [ "$REBUILD" = "1" ] || [ ! -f dist/desktop.html ]; }; then
    require_command npm

    if [ ! -d node_modules ]; then
        echo "==> Installing frontend dependencies"
        npm install --no-audit --no-fund
    fi

    echo "==> Building frontend"
    npm run build
fi

if [ "$BUILD_ONLY" = "1" ]; then
    echo "Build complete."
    exit 0
fi

if [ "$MODE" = "server" ]; then
    exec ./ratel-ledger server run
fi

if [ ! -f dist/desktop.html ]; then
    echo "Error: frontend build is missing. Run \"./run.sh --rebuild\"." >&2
    exit 1
fi

if [ ! -d electron/node_modules ]; then
    require_command npm
    echo "==> Installing Electron dependencies"
    (cd electron && npm install --no-audit --no-fund)
fi

if [ ! -x electron/node_modules/.bin/electron ]; then
    echo "Error: Electron is not installed. Run \"cd electron && npm install\"." >&2
    exit 1
fi

if [ "$(uname -s)" = "Linux" ] && [ -z "${DISPLAY:-}" ] && [ -z "${WAYLAND_DISPLAY:-}" ]; then
    echo "Error: no graphical session detected (DISPLAY and WAYLAND_DISPLAY are unset)." >&2
    echo "Use \"./run.sh --server\" to run only the backend." >&2
    exit 1
fi

cd electron

if [ "$(id -u)" = "0" ]; then
    exec ./node_modules/.bin/electron --no-sandbox .
fi

exec ./node_modules/.bin/electron .
