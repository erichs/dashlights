#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if fabric is installed
if ! command -v fabric >/dev/null 2>&1; then
    echo -e "${RED}❌ Error: fabric CLI not found${NC}"
    echo "Install it from: https://github.com/danielmiessler/fabric"
    exit 1
fi

# Check if fabric pattern is installed
if [ ! -f ~/.config/fabric/patterns/create_git_changelog/system.md ]; then
    echo -e "${YELLOW}⚠️  Fabric pattern not installed. Installing now...${NC}"
    make install-fabric-pattern
fi

# Check if working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    echo -e "${RED}❌ Error: Working directory is not clean${NC}"
    echo "Please commit or stash your changes before creating a release"
    git status --short
    exit 1
fi

# Get the latest tag
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
echo -e "${BLUE}Latest tag: ${LATEST_TAG}${NC}"

# Prompt for new version
echo ""
echo -e "${YELLOW}Enter new version (semver format, e.g., 0.4.0 or 1.0.0-alpha):${NC}"
read -r NEW_VERSION

# Validate semver format (supports pre-release versions)
# Format: X.Y.Z or X.Y.Z-prerelease (where prerelease can contain alphanumeric and hyphens)
if ! [[ "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?$ ]]; then
    echo -e "${RED}❌ Error: Invalid semver format.${NC}"
    echo -e "${RED}   Use X.Y.Z (e.g., 0.4.0) or X.Y.Z-prerelease (e.g., 1.0.0-alpha, 1.0.0-pre-release)${NC}"
    exit 1
fi

NEW_TAG="v${NEW_VERSION}"

# Check if tag already exists
if git rev-parse "$NEW_TAG" >/dev/null 2>&1; then
    echo -e "${RED}❌ Error: Tag ${NEW_TAG} already exists${NC}"
    exit 1
fi

# Get current date
RELEASE_DATE=$(date +%Y-%m-%d)

echo ""
echo -e "${BLUE}Generating changelog for ${NEW_TAG}...${NC}"

# Generate changelog content using fabric
CHANGELOG_CONTENT=$(git log "${LATEST_TAG}..HEAD" --oneline --no-merges | fabric -p create_git_changelog)

if [ -z "$CHANGELOG_CONTENT" ]; then
    echo -e "${RED}❌ Error: Failed to generate changelog content${NC}"
    exit 1
fi

# Create temporary file with new changelog entry
TEMP_CHANGELOG=$(mktemp)

# Write the new entry
echo "## [${NEW_VERSION}] - ${RELEASE_DATE}" > "$TEMP_CHANGELOG"
echo "" >> "$TEMP_CHANGELOG"
echo "$CHANGELOG_CONTENT" >> "$TEMP_CHANGELOG"
echo "" >> "$TEMP_CHANGELOG"

# If CHANGELOG.md exists, append the old content (skipping the header)
if [ -f CHANGELOG.md ]; then
    # Skip the first 6 lines (header) and append the rest
    tail -n +7 CHANGELOG.md >> "$TEMP_CHANGELOG"
else
    # Create header if CHANGELOG.md doesn't exist
    cat > CHANGELOG.md << 'EOF'
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

EOF
fi

# Add the header back
{
    echo "# Changelog"
    echo ""
    echo "All notable changes to this project will be documented in this file."
    echo ""
    echo "The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),"
    echo "and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html)."
    echo ""
    tail -n +1 "$TEMP_CHANGELOG"
} > CHANGELOG.md.new

mv CHANGELOG.md.new CHANGELOG.md
rm "$TEMP_CHANGELOG"

# Show the new changelog entry
echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}New changelog entry:${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
head -n 20 CHANGELOG.md | tail -n +7
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Prompt to continue
echo -e "${YELLOW}Create tag ${NEW_TAG} and commit CHANGELOG.md? (y/n):${NC}"
read -r CONFIRM

if [ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ]; then
    echo -e "${YELLOW}⚠️  Release cancelled. CHANGELOG.md has been updated but not committed.${NC}"
    exit 0
fi

# Commit the changelog
git add CHANGELOG.md
git commit -m "Update CHANGELOG for ${NEW_TAG}"

# Create the tag
git tag -a "$NEW_TAG" -m "Release ${NEW_TAG}"

echo ""
echo -e "${GREEN}✅ Release ${NEW_TAG} created successfully!${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo -e "  1. Review the changes: ${YELLOW}git show ${NEW_TAG}${NC}"
echo -e "  2. Push the commit: ${YELLOW}git push origin main${NC}"
echo -e "  3. Push the tag: ${YELLOW}git push origin ${NEW_TAG}${NC}"
echo ""
echo -e "${YELLOW}Or push both at once: git push origin main --tags${NC}"

