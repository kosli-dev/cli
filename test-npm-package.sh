#!/bin/bash
set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Testing NPM Package Locally ===${NC}\n"

# Check if kosli binary exists
if [ ! -f "./kosli" ]; then
    echo -e "${YELLOW}Building Kosli binary first...${NC}"
    make build
    echo -e "${GREEN}✓ Binary built${NC}\n"
fi

# Copy binary to npm-package structure
echo -e "${YELLOW}Setting up npm package structure...${NC}"
mkdir -p npm-package/bin
cp ./kosli npm-package/bin/
chmod +x npm-package/bin/kosli
echo -e "${GREEN}✓ Binary copied to npm-package/bin/${NC}\n"

# Extract version from kosli binary and update package.json
echo -e "${YELLOW}Syncing package version with kosli binary...${NC}"
KOSLI_VERSION=$(./kosli version 2>/dev/null | head -1 | grep -oP 'Version:"v\K[^"]+')
if [ -n "$KOSLI_VERSION" ]; then
    cd npm-package
    npm version "$KOSLI_VERSION" --no-git-tag-version --allow-same-version
    cd ..
    echo -e "${GREEN}✓ Package version set to ${KOSLI_VERSION}${NC}\n"
else
    echo -e "${RED}✗ Failed to extract version from kosli binary${NC}"
    exit 1
fi

# Pack the npm package
echo -e "${YELLOW}Packing npm package...${NC}"
cd npm-package

# Clean up any old package files before packing
rm -f kosli-cli-*.tgz

# Pack and capture the output filename
PACKAGE_FILE=$(npm pack 2>&1 | tail -1)
if [ -z "$PACKAGE_FILE" ] || [ ! -f "$PACKAGE_FILE" ]; then
    echo -e "${RED}✗ Failed to create package${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Package created: ${PACKAGE_FILE}${NC}\n"
cd ..

# Create test directory
TEST_DIR="/tmp/kosli-npm-test-$$"
echo -e "${YELLOW}Creating test directory: ${TEST_DIR}${NC}"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Install the package
echo -e "${YELLOW}Installing package...${NC}"
npm init -y > /dev/null 2>&1
npm install --no-save "$OLDPWD/npm-package/$PACKAGE_FILE"
echo -e "${GREEN}✓ Package installed${NC}\n"

# Test the CLI
echo -e "${YELLOW}Testing Kosli CLI...${NC}"
echo -e "${YELLOW}Running: npx kosli version${NC}"
npx kosli version
echo -e "${GREEN}✓ CLI executable${NC}\n"

# Test direct execution
echo -e "${YELLOW}Testing direct binary execution...${NC}"
./node_modules/.bin/kosli version > /dev/null 2>&1 || ./node_modules/@kosli/cli/bin/kosli version
echo -e "${GREEN}✓ Direct execution works${NC}\n"

# Cleanup
cd "$OLDPWD"
rm -rf "$TEST_DIR"
rm -f "npm-package/$PACKAGE_FILE"
echo -e "${GREEN}✓ Cleaned up test directory${NC}\n"

echo -e "${GREEN}=== All tests passed! ===${NC}"
echo -e "\nThe npm package is ready for publishing."
echo -e "To publish, run: ${YELLOW}npm publish --access public${NC}"
