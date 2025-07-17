#!/bin/bash

set -e

BUMP_TYPE=${1:-patch}  # patch, minor, or major
MODULES=("resagent" "resgate" "types")

bump_version() {
  local version=$1
  local bump=$2

  version="${version#v}"  # strip leading 'v'
  IFS='.' read -r major minor patch <<< "$version"

  case $bump in
    major)
      major=$((major + 1))
      minor=0
      patch=0
      ;;
    minor)
      minor=$((minor + 1))
      patch=0
      ;;
    patch)
      patch=$((patch + 1))
      ;;
    *)
      echo "Unknown bump type: $bump"
      exit 1
      ;;
  esac

  echo "v$major.$minor.$patch"
}

for module in "${MODULES[@]}"; do
  echo "Processing module: $module"

  # Find latest tag for this module
  latest_tag=$(git tag --sort=-v:refname | grep "^$module/v" | head -n 1)

  if [[ -z "$latest_tag" ]]; then
    echo "  No tag found, starting at v0.1.0"
    new_version="v0.1.0"
  else
    current_version="${latest_tag##*/}"  # Strip 'module/' part
    new_version=$(bump_version "$current_version" "$BUMP_TYPE")
    echo "  Latest: $current_version → New: $new_version"
  fi

  tag="${module}/${new_version}"
  git tag "$tag"
done

echo "All done. Push with:"
echo "  git push origin --tags"
