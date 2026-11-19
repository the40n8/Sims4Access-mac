#!/bin/zsh
# Bring changes from TS4 Access (the project SimSight is based on) into SimSight.
#
#   mac/sync-upstream.sh            merge upstream/dev into the current branch
#
# Steps: fetch upstream, merge, delete the Windows-only paths again
# (mac/windows-paths.txt, which also settles modify/delete conflicts on them),
# then report what is left to resolve by hand and whether a newer upstream
# release (UI package) exists than the one pinned in mac/ui-package.json.
set -euo pipefail
cd "${0:A:h:h}"

[[ -z "$(git status --porcelain)" ]] || { echo "Working tree not clean. Commit or stash first."; exit 1; }
git remote get-url upstream >/dev/null 2>&1 || \
  git remote add upstream https://github.com/LeonianUniverse478/Sims4Access.git
git fetch upstream --tags

BEFORE=$(git rev-parse HEAD)
NEW=$(git rev-list --count HEAD..upstream/dev)
if [[ "$NEW" == 0 ]]; then
  echo "Already up to date with upstream/dev."
else
  echo "Merging $NEW upstream commit(s):"
  git log --oneline HEAD..upstream/dev
  git merge --no-ff --no-commit upstream/dev || true

  # Re-apply the Windows removals (also resolves modify/delete conflicts).
  grep -v '^#' mac/windows-paths.txt | grep -v '^$' | while read -r spec; do
    git rm -r -q --ignore-unmatch --cached -- "$spec" 2>/dev/null || true
    for p in ${~spec}(N); do rm -rf -- "$p"; done
  done

  CONFLICTS=$(git diff --name-only --diff-filter=U)
  if [[ -n "$CONFLICTS" ]]; then
    echo
    echo "Resolve these by hand (keep upstream's change plus our macOS guard),"
    echo "then: python3 mac/rebrand.py && git add -A && git commit"
    echo "$CONFLICTS"
    exit 2
  fi
  # Upstream code uses the TS4 Access names (modules, commands, files):
  # rename it like ours. Changes to modules we already renamed merge through
  # git's rename detection; new modules are renamed here.
  python3 mac/rebrand.py
  git add -A
  git commit -q -m "Merge upstream/dev ($(git rev-parse --short upstream/dev)) into SimSight"
  echo "Merged. Changed upstream Python modules:"
  git diff --stat "$BEFORE" HEAD -- Scripts | tail -20
fi

echo
PINNED=$(python3 -I -c 'import json,sys; print(json.load(open(sys.argv[1]))["release"])' mac/ui-package.json)
LATEST=$(gh release list -R LeonianUniverse478/Sims4Access --limit 1 --json tagName --jq '.[0].tagName' 2>/dev/null || echo "?")
echo "UI package pinned: $PINNED   latest upstream release: $LATEST"
if [[ "$LATEST" != "?" && "$LATEST" != "$PINNED" ]]; then
  echo "A newer UI package exists. Check that its source matches what you merged, then: mac/pin-ui.sh $LATEST"
fi
echo "Next: mac/build.sh && mac/install.sh (game closed), then test in game."
