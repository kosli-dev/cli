#!/usr/bin/env bash
set -Eeu

SCRIPT_NAME="never_alone.sh"
PROPOSED_COMMIT=""
PULL_REQUEST_LIST_JSON_FILENAME=""
FAILED_PULL_REQUESTS_JSON_FILENAME=""


function print_help
{
    cat <<EOF
Use: $SCRIPT_NAME [options]

Script to get pull request info for all commits to main/master branch

Options are:
  -h                       Print this help menu
  -c <commit-sha>          Commit sha for release we are building now. Required
  -p <all-prs-filename>    Name of json file to save all pull-requests. Required
  -f <failed-prs-filename> Name of json file to save failed pull-requests: Required
EOF
}


function die
{
    echo "Error: $1" >&2
    exit 1
}


function repo_root
{
  git rev-parse --show-toplevel
}


function check_arguments
{
    while getopts "hc:p:f:" opt; do
        case $opt in
            h)
                print_help
                exit 1
                ;;
            c)
                PROPOSED_COMMIT=${OPTARG}
                ;;
            p)
                PULL_REQUEST_LIST_JSON_FILENAME=${OPTARG}
                ;;
            f)
                FAILED_PULL_REQUESTS_JSON_FILENAME=${OPTARG}
                ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2
                exit 1
                ;;
        esac
    done

    if [ -z "${PROPOSED_COMMIT}" ]; then
        die "option -c <commit-sha> is required"
    fi
    if [ -z "${PULL_REQUEST_LIST_JSON_FILENAME}" ]; then
        die "option -p <all-prs-filename> is required"
    fi
    if [ -z "${FAILED_PULL_REQUESTS_JSON_FILENAME}" ]; then
        die "option -f <failed-prs-filename> is required"
    fi
}

function main
{
    check_arguments "$@"

    # base_commit: the commit of latest release
    local -r base_commit=$($(repo_root)/bin/never_alone_get_commit_of_latest_release.sh)

    $(repo_root)/bin/never_alone_get_commits_with_pull_request_info.sh -b ${base_commit} -p ${PROPOSED_COMMIT} -o ${PULL_REQUEST_LIST_JSON_FILENAME}
    $(repo_root)/bin/never_alone_get_failing_pull_requests.sh -i ${PULL_REQUEST_LIST_JSON_FILENAME} -o ${FAILED_PULL_REQUESTS_JSON_FILENAME}
}

main "$@"
