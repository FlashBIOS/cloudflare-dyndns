#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

# Ensure we're on master branch.
current_branch=$(git symbolic-ref --short HEAD)
if [ "$current_branch" != "master" ]; then
  echo "Error: You must be on the master branch. You are currently on '$current_branch'." >&2
  exit 1
fi

# Check for uncommitted changes.
if ! git diff-index --quiet HEAD --; then
  echo "Error: There are uncommitted changes. Please commit or stash them before proceeding." >&2
  exit 1
fi

# Fetch the latest changes from the remote.
git fetch

# Check if the local master branch is up to date.
LOCAL=$(git rev-parse master)
REMOTE=$(git rev-parse origin/master)
if [ "$LOCAL" != "$REMOTE" ]; then
  echo "Error: The local master branch is not up to date with origin/master." >&2
  exit 1
fi

# Fetch all tags and find the latest semver tag
latest_tag=$(git tag --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -n 1)

if [[ -z "$latest_tag" ]]; then
  latest_tag="v0.0.0"
fi

printf "Latest tag: %s\n" "$latest_tag"

# Parse semver components
IFS='.' read -r -a parts <<< "${latest_tag#v}"
major=${parts[0]}
minor=${parts[1]}
patch=${parts[2]}

# Ask which part to bump
printf "What do you want to bump?\n"
options=("patch (${major}.${minor}.$((patch+1)))" "minor (${major}.$((minor+1)).0)" "major ($((major+1)).0.0)" "abort")
select bump in "${options[@]}"; do
  case $REPLY in
    1) next_tag="v${major}.${minor}.$((patch+1))"; break ;;
    2) next_tag="v${major}.$((minor+1)).0"; break ;;
    3) next_tag="v$((major+1)).0.0"; break ;;
    4) printf "Aborted\n"; exit 1 ;;
    *) printf "Invalid choice\n";;
  esac
done

printf "Next tag will be: %s\n" "$next_tag"

# Confirm and tag
read -rp "Tag this commit as $next_tag? [y/N] " confirm
if [[ "$confirm" == "y" || "$confirm" == "Y" ]]; then
  git tag "$next_tag"
  git push origin "$next_tag"
  printf "🚀 Tag %s pushed — release will be auto-created.\n" "$next_tag"
else
  printf "Aborted\n"
fi
