# Changeset

**Changeset** is a CLI tool for managing versioning and releases in multi-module Go repositories. It is inspired by [changesets](https://github.com/changesets/changesets), the excellent versioning and changelog management tool for JavaScript monorepos.

## Installation

```bash
go install github.com/gojekfarm/xtools/cmd/changeset@latest
```

## Commands

### `changeset init`

Sets up the `.changeset` folder with config and README.

```bash
$ changeset init
```

### `changeset` or `changeset add`

Create a changeset to document your changes.

```bash
$ changeset

? Which modules would you like to include?
  ◉ xkafka
  ◉ xkafka/middleware
  ○ xload

? Which modules should have a major bump?

? Which modules should have a minor bump?
  ◉ xkafka

? Summary:
> Added retry configuration to consumer

✓ Created .changeset/hungry-tiger-jump.md
```

**Flags:**

- `--empty` - Create an empty changeset (for changes that don't need releases)
- `--open` - Open the created changeset in your editor

### `changeset status`

Show pending changesets and computed version bumps.

```bash
$ changeset status

🦋  Changesets
   xkafka
     minor: hungry-tiger-jump
     patch: brave-lion-roar
```

**Flags:**

- `--verbose` - Show full changeset contents and release plan
- `--output=FILE` - Write JSON output for CI tools
- `--since=REF` - Only show changesets since a branch/tag

### `changeset version`

Consume changesets and update go.mod files.

```bash
$ changeset version

🦋  Consuming changesets
🦋  All files have been updated. Review changes and commit.
```

**Flags:**

- `--ignore=MODULE` - Skip specific modules from versioning
- `--snapshot` - Create snapshot versions for testing

### `changeset publish`

Create git tags and push to origin.

```bash
$ changeset publish

🦋  Publishing xkafka@v0.11.0
🦋  Publishing xkafka/middleware@v0.10.1
```

**Flags:**

- `--no-push` - Create tags locally without pushing

### `changeset tag`

Create git tags without pushing.

```bash
$ changeset tag

🦋  Creating tags
   xkafka/v0.11.0
   xkafka/middleware/v0.10.1
```

## Workflow

### Development (Feature Branch)

```
feature-branch
    │
    ├── Write code
    ├── changeset add              → creates .changeset/abc.md
    ├── git add -A
    ├── git commit -m "feat: ..."  → code + changeset committed together
    └── git push → Open PR
```

### Code Review (Pull Request)

```
PR #123: feature-branch → main
    │
    ├── Reviewers see code changes
    ├── Reviewers see .changeset/abc.md (bump types + summary)
    ├── CI runs tests
    └── Merge to main
```

### Accumulation (Main Branch)

```
main
    │
    ├── PR #123 merged → .changeset/abc.md
    ├── PR #124 merged → .changeset/def.md
    ├── PR #125 merged → .changeset/ghi.md
    │
    └── Changesets accumulate until release
```

### Release (Main Branch)

```
main
    │
    ├── changeset version
    │       ├── Reads all .changeset/*.md
    │       ├── Computes: xkafka v0.10.0 → v0.11.0
    │       ├── Updates go.mod files
    │       ├── Deletes .changeset/*.md
    │       └── Writes .changeset/release-manifest.json
    │
    ├── git add -A
    ├── git commit -m "chore: version packages"
    ├── git push
    │
    └── changeset publish
            ├── Reads release-manifest.json
            ├── Creates tags: xkafka/v0.11.0, ...
            ├── Pushes tags to origin
            └── Deletes release-manifest.json
```

### Result

```
Users can now:
    go get github.com/gojekfarm/xtools/xkafka@v0.11.0
```

## Configuration

`.changeset/config.json`:

```json
{
  "root": "github.com/gojekfarm/xtools",
  "baseBranch": "main",
  "ignore": ["cmd/*", "examples/*"],
  "dependentBump": "patch"
}
```

## Changeset File Format

```markdown
---
"xkafka": minor
"xkafka/middleware": patch
---

Added retry configuration to consumer.
```

## Release Manifest

The `version` and `publish` commands communicate via a manifest file.

### How It Works

```
changeset version
    │
    ├── Reads .changeset/*.md (changesets)
    ├── Computes version bumps
    ├── Updates go.mod files
    ├── Deletes consumed changesets
    └── Writes .changeset/release-manifest.json  ← created

changeset publish
    │
    ├── Reads .changeset/release-manifest.json
    ├── Creates git tags (xkafka/v0.11.0, etc.)
    ├── Pushes tags to origin
    └── Deletes .changeset/release-manifest.json ← cleaned up
```

### Manifest Format

`.changeset/release-manifest.json`:

```json
{
  "releases": [
    {
      "module": "xkafka",
      "version": "v0.11.0",
      "previousVersion": "v0.10.0",
      "bump": "minor"
    },
    {
      "module": "xkafka/middleware",
      "version": "v0.10.1",
      "previousVersion": "v0.10.0",
      "bump": "patch"
    },
    {
      "module": "xprom/xpromkafka",
      "version": "v0.10.1",
      "previousVersion": "v0.10.0",
      "bump": "patch",
      "reason": "dependency"
    }
  ]
}
```

| Field             | Description                                            |
| ----------------- | ------------------------------------------------------ |
| `module`          | Module short name (relative to root)                   |
| `version`         | New version to be tagged                               |
| `previousVersion` | Version before this release                            |
| `bump`            | Bump type: `major`, `minor`, or `patch`                |
| `reason`          | `"dependency"` if auto-bumped due to dependency change |

### Why a Manifest?

1. **Decouples version from publish** - You can review changes between steps
2. **Supports CI workflows** - Version in one job, publish in another
3. **Enables dry-run** - `version` can run without side effects to git
4. **Tracks intent** - Knows exactly what to tag without re-computing

## CI Automation

### 1. Require Changesets on PRs

Block PRs that modify code but don't include a changeset.

`.github/workflows/changeset-check.yml`:

```yaml
name: Changeset Check

on:
  pull_request:
    branches: [main]

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: Install changeset
        run: go install github.com/gojekfarm/xtools/cmd/changeset@latest

      - name: Check for changeset
        run: changeset status --since=origin/main
```

### 2. Automated Release PR

When changesets accumulate, automatically create a "Version Packages" PR.

`.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    branches: [main]

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: Install changeset
        run: go install github.com/gojekfarm/xtools/cmd/changeset@latest

      - name: Check for changesets
        id: check
        run: |
          if compgen -G ".changeset/*.md" > /dev/null; then
            echo "has_changesets=true" >> $GITHUB_OUTPUT
          else
            echo "has_changesets=false" >> $GITHUB_OUTPUT
          fi

      - name: Create Release PR
        if: steps.check.outputs.has_changesets == 'true'
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          # Configure git
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"

          # Create release branch
          BRANCH="changeset-release/main"
          git checkout -B $BRANCH

          # Version packages
          changeset version

          # Commit changes
          git add -A
          git commit -m "chore: version packages"
          git push -f origin $BRANCH

          # Create or update PR
          gh pr create --base main --head $BRANCH \
            --title "chore: version packages" \
            --body "This PR was auto-generated by the release workflow." \
            || gh pr edit $BRANCH --title "chore: version packages"
```

### 3. Publish on PR Merge

When the release PR is merged, publish tags.

`.github/workflows/publish.yml`:

```yaml
name: Publish

on:
  push:
    branches: [main]
    paths:
      - ".changeset/release-manifest.json"

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: Install changeset
        run: go install github.com/gojekfarm/xtools/cmd/changeset@latest

      - name: Configure git
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"

      - name: Publish
        run: changeset publish
```

### CI Flow Diagram

```
PR opened (feature branch)
    │
    └── changeset-check.yml
            │
            ├── Has .changeset/*.md? → ✓ Pass
            └── No changeset? → ✗ Fail

PR merged to main
    │
    └── release.yml
            │
            ├── Has changesets? → Create "Version Packages" PR
            └── No changesets? → Skip

"Version Packages" PR merged
    │
    └── publish.yml
            │
            ├── Has release-manifest.json? → Create & push tags
            └── No manifest? → Skip
```

### Skipping Changesets

For PRs that don't need releases (docs, CI config, etc.):

```bash
# Create an empty changeset
$ changeset add --empty
```

Or configure paths to ignore in `.changeset/config.json`:

```json
{
  "ignorePaths": ["*.md", ".github/**", "docs/**"]
}
```
