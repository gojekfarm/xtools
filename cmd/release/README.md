# Release

**Release** is a command line tool that automates the versioning and releasing of multi-module Go projects.

## Limitations

- All modules have the same version.
- Git tags are treated as the source of truth for the version.

## Usage

### List Modules

`list` command lists all the modules along with their versions & dependencies within the project.

```bash
$ release list
```

### Create Release

`create` command automates the release process for all modules in the repository. It performs the following steps:

- Checks for uncommitted changes and aborts if any are found.
- Determines the new version to use, either from `--version` or by auto-incrementing with `--major`, `--minor`, or `--patch` (only one may be specified).
- Updates the version in all `go.mod` files for every module in the repository.
- Updates inter-module dependencies so that modules within the repo reference the new version of each other.
- Writes all changes back to disk atomically, or aborts with an error if any step fails.

```bash
$ release create --version v0.1.0
# or auto-increment the version
$ release create --major | --minor | --patch
```

If any error occurs (such as invalid version, conflicting options, or file write failure), the process aborts and prints a clear error message.

### Tag

`tag` command creates Git tags for all the modules in the project. It reads the manifest file and creates Git tags for all the modules in the project.

```bash
# create tags
$ release tag
# or create and push the tags to the remote repository
$ release tag --push
```
