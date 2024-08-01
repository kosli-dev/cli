#!/usr/bin/env bash
set -Eeu

SCRIPT_NAME=get_commit_of_latest_release.sh

function die
{
    echo "Error: $1" >&2
    exit 1
}


function print_help
{
    cat <<EOF
Usage: $SCRIPT_NAME [options]

Script to get git commit for latest relased version of SW

Options are:
  -h               Print this help menu
EOF
}


function check_arguments
{
    while getopts "h" opt; do
        case $opt in
            h)
                print_help
                exit 1
                ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2
                exit 1
                ;;
        esac
    done
}


function get_commit_of_latest_relase
{
    latest_release=$(gh release list --exclude-pre-releases --exclude-drafts --limit 1 --json tagName,isLatest)
    if [ -z "$latest_release" -o "$latest_release" = "[]" ]; then
        die "Unable to get latest release"
    fi

    isLatest=$(echo "$latest_release" | jq ".[0].isLatest")
    if [ "$isLatest" != "true" ]; then
        die "Latest tag is not marked as 'isLatest'. $latest_release"
    fi
    latestTag=$(echo "$latest_release" | jq -r ".[0].tagName")
    git rev-list -n 1 ${latestTag}
}


function main {
    check_arguments "$@"
    get_commit_of_latest_relase
}

main "$@"
