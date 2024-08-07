#!/bin/bash

readonly SCRIPT_NAME=$(basename "$0")
readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

readonly SEMVER_REGEX="v(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)(\\-[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?"

usage() {
    cat <<- EOF
Usage: $SCRIPT_NAME [OPTIONS] SEMVER_TAG COMMIT_HASH

Creates git tag for all Go packages in project.

OPTIONS:
  -h --help        Show this help.

ARGUMENTS:
  SEMVER_TAG       Semantic version to tag with.
  COMMIT_HASH      Git commit hash to tag.
EOF
}

cmdline() {
    local arg commit

    for arg
    do
        local delim=""
        case "$arg" in
            # Translate long form options to short form.
            --help)           args="${args}-h ";;
            # Pass through for everything else.
            *) [[ "${arg:0:1}" == "-" ]] || delim="\""
                args="${args}${delim}${arg}${delim} ";;
        esac
    done

    # Reset and process short form options.
    eval set -- "$args"

    while getopts "h" OPTION
    do
         case $OPTION in
         h)
             usage
             exit 0
             ;;
         *)
             echo "unknown option: $OPTION"
             usage
             exit 1
             ;;
        esac
    done

    # Positional arguments.
    shift $((OPTIND-1))
    readonly TAG="$1"
    if [ -z "$TAG" ]
    then
        echo "missing SEMVER_TAG"
        usage
        exit 1
    fi
    if [[ ! "$TAG" =~ $SEMVER_REGEX ]]
    then
        printf "invalid semantic version: %s\n" "$TAG"
        exit 2
    fi
    if [[ "$( git tag --list "$TAG" )" ]]
    then
        printf "tag already exists: %s\n" "$TAG"
        exit 2
    fi

    shift
    commit="$1"
    if [ -z "$commit" ]
    then
        echo "missing COMMIT_HASH"
        usage
        exit 1
    fi
    # Verify rev is for a commit and unify hashes into a complete SHA1.
    readonly SHA="$( git rev-parse --quiet --verify "${commit}^{commit}" )"
    if [ -z "$SHA" ]
    then
        printf "invalid commit hash: %s\n" "$commit"
        exit 2
    fi
    if [ "$( git merge-base "$SHA" HEAD )" != "$SHA" ]
    then
        printf "commit '%s' not found on this branch\n" "$commit"
        exit 2
    fi
}

package_dirs() {
    # Return a list of package directories in the form:
    #
    #  package/directory/a
    #  package/directory/b
    #  deeper/package/directory/a
    #  ...
    #
    find . -mindepth 2 -type f -name 'go.mod' -exec dirname {} \; \
        | sed 's/^\.\///' \
        | sort
}

git_tag() {
    local tag="$1"
    local commit="$2"

    git tag -a "$tag" -s -m "Version $tag" "$commit"
}

previous_version() {
    local current="$1"

    # Requires git > 2.0
    git tag -l --sort=v:refname \
        | grep -E "^${SEMVER_REGEX}$" \
        | grep -v "$current" \
        | tail -1
}

print_changes() {
    local tag="$1"
    local previous

    previous="$( previous_version "$tag" )"
    if [ -n "$previous" ]
    then
        printf "\nRaw changes made between %s and %s\n" "$previous" "$tag"
        printf "======================================\n"
        git --no-pager log --pretty=oneline "${previous}..$tag"
    fi
}

main() {
    local dir

    cmdline "$@"

    cd "$SCRIPT_DIR" || exit 3

    # Create tag for root package.
    git_tag "$TAG" "$SHA"
    printf "created tag: %s\n" "$TAG"

    # Create tag for all sub-packages.
    for dir in $( package_dirs )
    do
        git_tag "${dir}/$TAG" "$SHA"
        printf "created tag: %s\n" "${dir}/$TAG"        
    done

    printf "\nPushing tags to remote\n"

    for dir in $( package_dirs )
    do
        git push origin "${dir}/$TAG"
    done

    print_changes "$TAG"
}

main "$@"
