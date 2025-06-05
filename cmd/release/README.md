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

### Build Manifest

`manifest` command builds the manifest file for the project. It crawls the project, finds all modules and their dependencies, and reads the respective Git tags to determine the version of each module.

The manifest file contains the following information:

- Module path
- Current version
- Dependencies and their versions (within the same repository)

_Note: Manifest file is temporary and is deleted after the release is created._

```bash
$ release manifest
```

### Create Release

`create` command creates a release branch and updates the version of the project. It reads the manifest file and updates the version of the project and all the modules within the same repository. It updates the new version in the manifest file.

```bash
$ release create --version v0.1.0
# or auto-increment the version
$ release create --major | --minor | --patch
```

### Tag

`tag` command creates Git tags for all the modules in the project. It reads the manifest file and creates Git tags for all the modules in the project.

```bash
# create tags
$ release tag
# or create and push the tags to the remote repository
$ release tag --push
```
