#!/bin/sh
set -eu

# This script downloads the OS- and architecture-specific Kosli CLI binary,
# extracts it, and moves the executable to a directory in your PATH.

# --- Configuration ---
CLI_OS="unknown"
ARCH="unknown"
VERSION=""
FILE_NAME="kosli"
DEBUG=false

# --- Debug function ---
debug_print() {
    if [ "$DEBUG" = true ]; then
        echo "DEBUG: $1" >&2
    fi
}

# --- Parse arguments ---
while [ $# -gt 0 ]; do
    case $1 in
        --debug)
            DEBUG=true
            debug_print "Debug mode enabled"
            shift
            ;;
        *)
            VERSION=$1
            debug_print "Version specified: $VERSION"
            shift
            ;;
    esac
done

# --- Version Selection ---
if [ -n "$VERSION" ]; then
    echo "Downloading specified version $VERSION of Kosli CLI..."
    debug_print "Using specified version: $VERSION"
else
    echo "Detecting the latest version of Kosli CLI..."
    debug_print "Fetching latest version from GitHub API"
    # Fetches the latest release tag from the GitHub API
    METADATA=$(curl -s "https://api.github.com/repos/kosli-dev/cli/releases/latest")
    debug_print "GitHub API response: $METADATA"
    TAG_NAME=$(echo "$METADATA" | grep '"tag_name":')
    debug_print "GitHub API response tag: $TAG_NAME"
    LATEST_TAG=$(echo "$TAG_NAME" | sed -E 's/.*"([^"]+)".*/\1/')
    debug_print "GitHub API response tag: $LATEST_TAG"
    if [ -z "$LATEST_TAG" ]; then
        echo "Error: Could not fetch the latest version tag from GitHub."
        exit 1
    fi
    VERSION=$LATEST_TAG
    echo "Latest version is $VERSION. Downloading..."
    debug_print "Set VERSION to: $VERSION"
fi
echo ""

# Strip the 'v' prefix for use in the filename, e.g., v2.11.22 -> 2.11.22
VERSION_FILENAME=$(echo "$VERSION" | sed 's/^v//')
debug_print "VERSION_FILENAME after stripping 'v': $VERSION_FILENAME"

# --- OS and Architecture Detection ---
debug_print "Detecting OS and architecture"
debug_print "uname -s output: $(uname -s)"
debug_print "uname -m output: $(uname -m)"

UNAME_S=$(uname -s)
if echo "$UNAME_S" | grep -q -E -i "(cygwin|mingw|msys|windows)"; then
    CLI_OS="windows"
    ARCH="amd64"
    FILE_NAME="${FILE_NAME}.exe"
    debug_print "Detected Windows OS"
elif echo "$UNAME_S" | grep -q -i "darwin"; then
    CLI_OS="darwin"
    debug_print "Detected Darwin/macOS"
    UNAME_M=$(uname -m)
    if [ "$UNAME_M" = "arm64" ]; then
      ARCH="arm64"
      debug_print "Detected ARM64 architecture"
    else
      ARCH="amd64"
      debug_print "Detected AMD64 architecture"
    fi
else
    CLI_OS="linux"
    debug_print "Detected Linux OS"
    MACHINE_TYPE="$(uname -m)"
    debug_print "Machine type: $MACHINE_TYPE"
    case $MACHINE_TYPE in
        amd64 | x86_64 | x64)
            ARCH="amd64"
            debug_print "Mapped to AMD64 architecture"
            ;;
        aarch64 | arm64)
            ARCH="arm64"
            debug_print "Mapped to ARM64 architecture"
            ;;
        *)
            echo "Error: Unsupported Linux architecture: $MACHINE_TYPE"
            echo "Kosli CLI is only available for amd64 and arm64 on Linux."
            exit 1
            ;;
    esac
fi

debug_print "Final values - CLI_OS: $CLI_OS, ARCH: $ARCH, FILE_NAME: $FILE_NAME"

# --- Download and Extract ---
# The download is a .tar.gz or .zip file which needs to be extracted
if [ "$CLI_OS" = "windows" ]; then
    URL="https://github.com/kosli-dev/cli/releases/download/${VERSION}/kosli_${VERSION_FILENAME}_${CLI_OS}_${ARCH}.zip"
    debug_print "Windows URL constructed: $URL"
    echo "Downloading from: $URL"
    # Download and extract for Windows
    debug_print "Starting Windows download and extraction"
    if ! curl -L --fail "$URL" -o kosli.zip; then
        echo "Error: Download failed. Please check the URL and your network connection."
        exit 1
    fi
    debug_print "Download completed, extracting zip file"
    unzip -o kosli.zip
    debug_print "Extraction completed"
else
    URL="https://github.com/kosli-dev/cli/releases/download/${VERSION}/kosli_${VERSION_FILENAME}_${CLI_OS}_${ARCH}.tar.gz"
    debug_print "Unix URL constructed: $URL"
    echo "Downloading from: $URL"
    # Download and extract for Linux and Darwin
    debug_print "Starting Unix download and extraction"
    if ! curl -L --fail "$URL" | tar zx; then
        echo "Error: Download or extraction failed. Please check the URL and your network connection."
        exit 1
    fi
    debug_print "Download and extraction completed"
fi

# --- Installation ---
# Move the extracted binary to a directory in the user's PATH
echo "Installing Kosli CLI..."
debug_print "Starting installation process"
debug_print "Current PATH: $PATH"

# Check directories one by one instead of using set --
for dir in "/usr/local/bin" "/usr/bin" "/opt/bin"; do
    debug_print "Checking directory: $dir"
    # Check if destination directory exists and is in the PATH
    if [ -d "$dir" ] && echo "$PATH" | grep -q "$dir"; then
        debug_print "Directory $dir exists and is in PATH"
        debug_print "Attempting to move $FILE_NAME to $dir"
        if mv "$FILE_NAME" "$dir/"; then
            echo ""
            echo "✅ Kosli CLI was successfully installed in $dir"
            echo "Running 'kosli version' to verify:"
            debug_print "Installation successful, running version check"
            kosli version
            exit 0
        else
            echo ""
            echo "Attempting to install with sudo..."
            echo "We'd like to install the Kosli CLI executable in '$dir'. Please enter your password if prompted."
            debug_print "Regular move failed, trying with sudo"
            if sudo mv "$FILE_NAME" "$dir/"; then
                echo ""
                echo "✅ Kosli CLI was successfully installed in $dir"
                echo "Running 'kosli version' to verify:"
                debug_print "Sudo installation successful, running version check"
                kosli version
                exit 0
            fi
            debug_print "Sudo move also failed for $dir"
        fi
    else
        debug_print "Directory $dir either doesn't exist or is not in PATH"
    fi
done

debug_print "All installation attempts failed"
echo ""
echo "Error: Could not install Kosli CLI."
echo "Please move the '$FILE_NAME' executable manually to a directory in your \$PATH."
echo "For example, you can run: sudo mv \"$FILE_NAME\" /usr/local/bin/"
exit 1