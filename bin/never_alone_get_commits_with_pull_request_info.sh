#!/usr/bin/env bash
set -Eeu

SCRIPT_NAME=never_alone_get_commits_with_pull_reqeust_info.sh
OUTPUT_FILE=""
BASE_COMMIT=""
PROPOSED_COMMIT=""


function print_help
{
    cat <<EOF
Use: $SCRIPT_NAME [options]

Script that gets all commits between base_commit and proposed_commit
and collects pull-request information about them. Stores the results in a json file.

Options are:
  -h                   Print this help menu
  -b <base-commit>     Oldest commit sha. Required
  -p <proposed-commit> Newest commit sha. Required
  -o <output-file>     Result output file. Required
EOF
}


function die
{
    echo "Error: $1" >&2
    exit 1
}


function check_arguments
{
    while getopts "hb:o:p:" opt; do
        case $opt in
            h)
                print_help
                exit 1
                ;;
            b)
                BASE_COMMIT=${OPTARG}
                ;;
            o)
                OUTPUT_FILE=${OPTARG}
                ;;
            p)
                PROPOSED_COMMIT=${OPTARG}
                ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2
                exit 1
                ;;
        esac
    done

    if [ -z "${BASE_COMMIT}" ]; then
        die "option -b <base-commit> is required"
    fi
    if [ -z "${OUTPUT_FILE}" ]; then
        die "option -o <output-file> is required"
    fi
    if [ -z "${PROPOSED_COMMIT}" ]; then
        die "option -p <proposed-commit> is required"
    fi
}


function get_pull_requests
{
    local base_commit=$1; shift
    local proposed_commit=$1; shift
    local result_file=$1; shift
    local commits list_separator
    # Use gh instead of git so we can keep the commit depth of 1. The order of the response for gh is reversed
    # so I do a tac at the end to get it the same order.
    commits=($(gh api repos/:owner/:repo/compare/${base_commit}...${proposed_commit} -q '.commits[].sha' | tac))

    list_separator=""
    echo "[" > ${result_file}
    for commit in "${commits[@]}"; do
        echo "${list_separator}" >> ${result_file}
        pr_data=$(gh pr list --search "${commit}" --state merged --json author,reviews,mergeCommit,mergedAt,reviewDecision,url)
        if [ "$pr_data" = "[]" ]; then
            # Commit is not merged back to master (this will happen if you run this on a branch)
            echo '{"sha": "'$commit'"}' >> "${result_file}"
        elif [ "$(echo "$pr_data" | jq '. | length')" -eq 0 ]; then
            # No pull request found for that commit, so do a new request to get the commit
            commit_data=$(gh search commits --hash "${commit}" --json author,sha)
            echo "$commit_data" | jq '.[0]' >> "${result_file}"
        else
            # The PR data does not contain the commit sha so we add it manually. Use 'sha' as key since that is
            # what is used in the 'gh search commits' command
            echo "$pr_data" | jq --arg sha "$commit" '.[] | . + {sha: $sha}' >> "${result_file}"
        fi
        list_separator=","
    done
    echo "]" >> ${result_file}
}


function main
{
    check_arguments "$@"
    get_pull_requests ${BASE_COMMIT} ${PROPOSED_COMMIT} ${OUTPUT_FILE}
}

main "$@"
