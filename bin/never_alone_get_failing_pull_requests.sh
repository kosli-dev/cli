#!/usr/bin/env bash
set -Eeu

SCRIPT_NAME=get_failing_pull_requests.sh
INPUT_FILE=""
OUTPUT_FILE=""


function print_help
{
    cat <<EOF
Usage: $SCRIPT_NAME [options]

Script to parse pull request info file to check that all commits have pull-request with committer != approver.
Intended to run on the output of the `never_alone_get_commits_with_pull_request_info.sh` script

Options are:
  -h               Print this help menu
  -i <input-file>  Json input file. Required
  -o <output-file> Output file. Required
EOF
}


function die
{
    echo "Error: $1" >&2
    exit 1
}


function check_arguments
{
    while getopts "hi:o:" opt; do
        case $opt in
            h)
                print_help
                exit 1
                ;;
            i)
                INPUT_FILE=${OPTARG}
                ;;
            o)
                OUTPUT_FILE=${OPTARG}
                ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2
                exit 1
                ;;
        esac
    done

    if [ -z "${INPUT_FILE}" ]; then
        die "option -i <input-file> is required"
    fi
    if [ -z "${OUTPUT_FILE}" ]; then
        die "option -o <output-file> is required"
    fi
}


function get_failing_pull_requests
{
    local file=$1;shift
    local failed_reviews=()

    # Read each entry and check it
    while IFS= read -r entry; do
        # Check for missing latestReviews or if that list is empty
        latest_reviews=$(echo "$entry" | jq '.latestReviews // empty')
        if [ -z "$latest_reviews" -o "$latest_reviews" = "[]" ]; then
            failed_reviews+=("$entry")
        else
            # Find the entry where 'state' is APPROVED
            commit_author=$(echo "$entry" | jq '.author.login')
            reviews_length=$(echo "$entry" | jq '.latestReviews | length')
            for i in $(seq 0 $(( reviews_length - 1 )))
            do
                review=$(echo "$entry" | jq ".latestReviews[$i]")
                state=$(echo "$review" | jq ".state")
                if [ "$state" = '"APPROVED"' ]; then
                    break
                fi
            done
            if [ "$state" != '"APPROVED"' ]; then
                # Fail if no APPROVED was found
                failed_reviews+=("$entry")
            else
                # Fail if latest reviewer and auther is the same person
                review_author=$(echo "$review" | jq ".author.login")
                if [ "${review_author}" = "${commit_author}" ]; then
                    failed_reviews+=("$entry")
                fi
            fi
        fi
    done < <(jq -c '.[]' "$file")

    echo "${failed_reviews[@]}" | jq  -s '.' > ${OUTPUT_FILE}
}


function main
{
    check_arguments "$@"
    get_failing_pull_requests ${INPUT_FILE}
}

main "$@"
