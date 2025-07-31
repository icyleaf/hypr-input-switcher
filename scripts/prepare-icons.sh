#!/bin/bash

set -e

ICONS_DIR="internal/notification/icons"
CIRCLE_FLAGS_REPO="https://github.com/HatScripts/circle-flags.git"

echo "Preparing icons for embedding..."

# Create icons directory
mkdir -p $ICONS_DIR

# If directory is empty, download icons
if [ ! "$(ls -A $ICONS_DIR 2>/dev/null)" ]; then
    echo "Downloading circle-flags icons..."

    # Temporary directory
    TEMP_DIR=$(mktemp -d)

    # Clone repository (only latest version)
    git clone --depth=1 $CIRCLE_FLAGS_REPO $TEMP_DIR

    # Copy needed icons
    COUNTRIES=("us" "cn" "jp" "kr" "de" "fr" "es" "ru" "sa" "in" "gb" "it" "br" "mx" "th" "vn" "hk")

    for country in "${COUNTRIES[@]}"; do
        if [ -f "$TEMP_DIR/flags/$country.svg" ]; then
            cp "$TEMP_DIR/flags/$country.svg" "$ICONS_DIR/"
            echo "✓ Copied $country.svg"
        else
            echo "✗ Not found: $country.svg"
        fi
    done

    cp "$TEMP_DIR/LICENSE.md" "$ICONS_DIR/"

    # Clean up temporary directory
    rm -rf $TEMP_DIR

    echo "Icons prepared successfully! ($(ls $ICONS_DIR | wc -l) files)"
else
    echo "Icons already exist ($(ls $ICONS_DIR | wc -l) files), skipping download."
fi
