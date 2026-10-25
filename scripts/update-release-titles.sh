#!/usr/bin/env bash
set -euo pipefail

# Check if gh cli is installed
if ! command -v gh &> /dev/null; then
    echo "Error: gh command not found. Please install GitHub CLI." >&2
    exit 1
fi

echo "Fetching releases..."
releases=$(gh release list --limit 100 --json tagName,name)

echo "$releases" | jq -c '.[]' | while read -r item; do
    tag=$(echo "$item" | jq -r '.tagName')
    name=$(echo "$item" | jq -r '.name')

    # Remove 'hypr-input-switcher ' prefix if present
    new_name=$(echo "$name" | sed -E 's/^hypr-input-switcher[[:space:]]+//')

    if [ "$name" != "$new_name" ]; then
        echo "Updating release $tag: '$name' -> '$new_name'..."
        gh release edit "$tag" --title "$new_name"
    else
        echo "Release $tag already has title: '$name', skipping."
    fi
done

echo "Done updating releases!"
