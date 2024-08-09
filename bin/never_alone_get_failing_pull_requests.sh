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

    # Read each pull-request entry and check it
    while IFS= read -r pr_data; do
        # Check for missing reviews or if that list is empty
        reviews=$(echo "${pr_data}" | jq '.[0].reviews')
        github_review_decision=$(echo "${pr_data}" | jq '.[0].reviewDecision')
        local compliant="false"
        if [ "$reviews" = "null" ]; then
            pr_data=$(echo $pr_data | jq '. += {"failure": "no pull-request"}')
            failed_reviews+=("$pr_data")
        elif [ -z "$reviews" -o "$reviews" = "[]" ]; then
            pr_data=$(echo $pr_data | jq '. += {"failure": "no reviewers"}')
            failed_reviews+=("$pr_data")
        elif [ "${github_review_decision}" != '"APPROVED"' ]; then
            pr_data=$(echo $pr_data | jq '. += {"failure": "pull-request not approved"}')
            failed_reviews+=("$pr_data")
        else
            # Loop over reviews and check that at least one approver is not the same as committer
            pr_author=$(echo "${pr_data}" | jq '.[0].author.login')
            reviews_length=$(echo "${pr_data}" | jq '.[0].reviews | length')
            for i in $(seq 0 $(( reviews_length - 1 )))
            do
                review=$(echo "${pr_data}" | jq ".[0].reviews[$i]")
                state=$(echo "$review" | jq ".state")
                review_author=$(echo "$review" | jq ".author.login")
                if [ "$state" = '"APPROVED"' -a "${review_author}" != "${pr_author}" ]; then
                    compliant="true"
                fi
            done

            if [ "${compliant}" == "false" ]; then
                pr_data=$(echo $pr_data | jq '. += {"failure": "committer and approver are the same person"}')
                failed_reviews+=("$pr_data")
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
