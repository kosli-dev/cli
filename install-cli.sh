#!/bin/sh
set -eu

# This script downloads the OS- and architecture-specific Kosli CLI binary,
# extracts it, and moves the executable to a directory in your PATH.

# --- Configuration ---
CLI_OS="unknown"
ARCH="unknown"
VERSION=""
FILE_NAME="kosli"

# --- Version Selection ---
if [ $# -eq 1 ]; then
    VERSION=$1
    echo "Downloading specified version $VERSION of Kosli CLI..."
else
    echo "Detecting the latest version of Kosli CLI..."
    # Fetches the latest release tag from the GitHub API
    LATEST_TAG=$(curl -s "https://api.github.com/repos/kosli-dev/cli/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$LATEST_TAG" ]; then
        echo "Error: Could not fetch the latest version tag from GitHub."
        exit 1
    fi
    VERSION=$LATEST_TAG
    echo "Latest version is $VERSION. Downloading..."
fi
echo ""

# Strip the 'v' prefix for use in the filename, e.g., v2.11.22 -> 2.11.22
VERSION_FILENAME=$(echo "$VERSION" | sed 's/^v//')

# --- OS and Architecture Detection ---
if uname -s | grep -q -E -i "(cygwin|mingw|msys|windows)"; then
    CLI_OS="windows"
    ARCH="amd64"
    FILE_NAME="${FILE_NAME}.exe"
elif uname -s | grep -q -i "darwin"; then
    CLI_OS="darwin"
    if [ "$(uname -m)" = "arm64" ]; then
      ARCH="arm64"
    else
      ARCH="amd64"
    fi
else
    CLI_OS="linux"
    MACHINE_TYPE="$(uname -m)"
    case $MACHINE_TYPE in
        amd64 | x86_64 | x64)
            ARCH="amd64"
            ;;
        aarch64 | arm64)
            ARCH="arm64"
            ;;
        *)
            echo "Error: Unsupported Linux architecture: $MACHINE_TYPE"
            echo "Kosli CLI is only available for amd64 and arm64 on Linux."
            exit 1
            ;;
    esac
fi

# --- Download and Extract ---
# The download is a .tar.gz or .zip file which needs to be extracted
if [ "$CLI_OS" = "windows" ]; then
    URL="https://github.com/kosli-dev/cli/releases/download/${VERSION}/kosli_${VERSION_FILENAME}_${CLI_OS}_${ARCH}.zip"
    # Download and extract for Windows
    if ! curl -L --fail "$URL" -o kosli.zip; then
        echo "Error: Download failed. Please check the URL and your network connection."
        exit 1
    fi
    unzip -o kosli.zip
else
    URL="https://github.com/kosli-dev/cli/releases/download/${VERSION}/kosli_${VERSION_FILENAME}_${CLI_OS}_${ARCH}.tar.gz"
    # Download and extract for Linux and Darwin
    if ! curl -L --fail "$URL" | tar zx; then
        echo "Error: Download or extraction failed. Please check the URL and your network connection."
        exit 1
    fi
fi

# --- Installation ---
# Move the extracted binary to a directory in the user's PATH
echo "Installing Kosli CLI..."
set -- "/usr/local/bin" "/usr/bin" "/opt/bin"
while [ -n "$1" ]; do
    # Check if destination directory exists and is in the PATH
    if [ -d "$1" ] && echo "$PATH" | grep -q "$1"; then
        if mv "$FILE_NAME" "$1/"; then
            echo ""
            echo "✅ Kosli CLI was successfully installed in $1"
            echo "Running 'kosli version' to verify:"
            kosli version
            exit 0
        else
            echo ""
            echo "Attempting to install with sudo..."
            echo "We'd like to install the Kosli CLI executable in '$1'. Please enter your password if prompted."
            if sudo mv "$FILE_NAME" "$1/"; then
                echo ""
                echo "✅ Kosli CLI was successfully installed in $1"
                echo "Running 'kosli version' to verify:"
                kosli version
                exit 0
            fi
        fi
    fi
    shift
done

echo ""
echo "Error: Could not install Kosli CLI."
echo "Please move the '$FILE_NAME' executable manually to a directory in your \$PATH."
echo "For example, you can run: sudo mv \"$FILE_NAME\" /usr/local/bin/"
exit 1