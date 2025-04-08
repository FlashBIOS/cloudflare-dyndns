#!/bin/bash
set -e

# Fetch all tags and find the latest semver tag
latest_tag=$(git tag --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -n 1)

if [[ -z "$latest_tag" ]]; then
  latest_tag="v0.0.0"
fi

echo "Latest tag: $latest_tag"

# Parse semver components
IFS='.' read -r -a parts <<< "${latest_tag#v}"
major=${parts[0]}
minor=${parts[1]}
patch=${parts[2]}

# Ask which part to bump
echo "What do you want to bump?"
select bump in "patch (${major}.${minor}.$((patch+1)))" \
                "minor (${major}.$((minor+1)).0)" \
                "major ($((major+1)).0.0)" "abort"; do
  case $REPLY in
    1) next_tag="v${major}.${minor}.$((patch+1))"; break ;;
    2) next_tag="v${major}.$((minor+1)).0"; break ;;
    3) next_tag="v$((major+1)).0.0"; break ;;
    4) echo "Aborted"; exit 1 ;;
    *) echo "Invalid choice";;
  esac
done

echo "Next tag will be: $next_tag"

# Confirm and tag
read -p "Tag this commit as $next_tag? [y/N] " confirm
if [[ "$confirm" == "y" || "$confirm" == "Y" ]]; then
  git tag "$next_tag"
  git push origin "$next_tag"
  echo "🚀 Tag $next_tag pushed — release will be auto-created."
else
  echo "Aborted"
fi