# Changesets

This directory contains changeset files that describe changes to the codebase.

## What is a changeset?

A changeset is a file that describes which packages should be released and how
(major, minor, or patch). When it's time to release, these changesets are
consumed to determine version bumps.

## Creating a changeset

Run:
```
changeset add
```

This will interactively create a changeset file.

## File format

```markdown
---
"package-name": minor
"other-package": patch
---

Description of the changes.
```
