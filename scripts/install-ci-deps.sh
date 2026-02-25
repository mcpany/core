#!/bin/sh
set -e
set -x

echo "Installing CI dependencies..."
echo "User: $(id -u):$(id -g)"

# Check if dependencies are already installed
MISSING_DEPS=""
for arg in "$@"; do
    case "$arg" in
        xz-utils) CMD="xz" ;;
        build-essential) CMD="make" ;;
        ca-certificates) CMD="update-ca-certificates" ;; # Rough check, or check file existence?
        *) CMD="$arg" ;;
    esac

    # Special handling for ca-certificates check?
    if [ "$arg" = "ca-certificates" ]; then
        if [ ! -f /etc/ssl/certs/ca-certificates.crt ] && [ ! -f /etc/pki/tls/certs/ca-bundle.crt ]; then
             MISSING_DEPS="$MISSING_DEPS $arg"
        fi
    elif ! command -v "$CMD" >/dev/null 2>&1; then
        MISSING_DEPS="$MISSING_DEPS $arg"
    fi
done

if [ -z "$MISSING_DEPS" ]; then
    echo "All dependencies already installed."
    exit 0
fi

echo "Installing missing dependencies: $MISSING_DEPS"
# Reset args to missing deps
set -- $MISSING_DEPS

SUDO=""
if command -v sudo >/dev/null 2>&1; then
    SUDO="sudo"
fi

install_apt() {
    echo "Detected apt-get."
    if [ -n "$SUDO" ]; then
        $SUDO apt-get update || true
        $SUDO DEBIAN_FRONTEND=noninteractive apt-get install -y "$@"
    else
        apt-get update || true
        DEBIAN_FRONTEND=noninteractive apt-get install -y "$@"
    fi
}

install_apk() {
    echo "Detected apk."
    if [ -n "$SUDO" ]; then
        $SUDO apk add --no-cache "$@"
    else
        apk add --no-cache "$@"
    fi
}

install_dnf() {
    echo "Detected dnf."
    if [ -n "$SUDO" ]; then
        $SUDO dnf install -y "$@"
    else
        dnf install -y "$@"
    fi
}

install_yum() {
    echo "Detected yum."
    if [ -n "$SUDO" ]; then
        $SUDO yum install -y "$@"
    else
        yum install -y "$@"
    fi
}

install_microdnf() {
    echo "Detected microdnf."
    if [ -n "$SUDO" ]; then
        $SUDO microdnf install -y "$@"
    else
        microdnf install -y "$@"
    fi
}

install_zypper() {
    echo "Detected zypper."
    if [ -n "$SUDO" ]; then
        $SUDO zypper install -y "$@"
    else
        zypper install -y "$@"
    fi
}

# Determine package manager
if command -v apt-get >/dev/null 2>&1; then
    install_apt "$@"
elif command -v apk >/dev/null 2>&1; then
    install_apk "$@"
elif command -v dnf >/dev/null 2>&1; then
    # Map packages if needed?
    # xz-utils -> xz
    # build-essential -> make
    # But we pass args directly.
    # We should normalize args before calling.
    NEW_ARGS=""
    for arg in "$@"; do
        case "$arg" in
            xz-utils) NEW_ARGS="$NEW_ARGS xz" ;;
            build-essential) NEW_ARGS="$NEW_ARGS make" ;;
            *) NEW_ARGS="$NEW_ARGS $arg" ;;
        esac
    done
    install_dnf $NEW_ARGS
elif command -v yum >/dev/null 2>&1; then
    NEW_ARGS=""
    for arg in "$@"; do
        case "$arg" in
            xz-utils) NEW_ARGS="$NEW_ARGS xz" ;;
            build-essential) NEW_ARGS="$NEW_ARGS make" ;;
            *) NEW_ARGS="$NEW_ARGS $arg" ;;
        esac
    done
    install_yum $NEW_ARGS
elif command -v microdnf >/dev/null 2>&1; then
    NEW_ARGS=""
    for arg in "$@"; do
        case "$arg" in
            xz-utils) NEW_ARGS="$NEW_ARGS xz" ;;
            build-essential) NEW_ARGS="$NEW_ARGS make" ;;
            *) NEW_ARGS="$NEW_ARGS $arg" ;;
        esac
    done
    install_microdnf $NEW_ARGS
elif command -v zypper >/dev/null 2>&1; then
    NEW_ARGS=""
    for arg in "$@"; do
        case "$arg" in
            xz-utils) NEW_ARGS="$NEW_ARGS xz" ;;
            build-essential) NEW_ARGS="$NEW_ARGS make" ;;
            *) NEW_ARGS="$NEW_ARGS $arg" ;;
        esac
    done
    install_zypper $NEW_ARGS
else
    echo "No supported package manager found."
    exit 1
fi

echo "Dependencies installed successfully."
