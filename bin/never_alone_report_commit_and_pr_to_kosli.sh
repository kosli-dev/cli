#!/usr/bin/env bash
set -Eeu

SCRIPT_NAME=never_alone_report_commit_and_pr_to_kosli.sh
BASE_COMMIT=""
PROPOSED_COMMIT=""
FLOW_NAME=""
TRAIL_NAME=""


function print_help
{
    cat <<EOF
Use: $SCRIPT_NAME [options]

Script that gets commit and pull-request info for a commit sha and report them to kosli

Options are:
  -h                   Print this help menu
  -b <base-commit>     Oldest commit sha. Required
  -p <proposed-commit> Newest commit sha. Required
  -f <flow-name>       Flow name to report commit and pull-request info. Required
  -t <trail-name>      Name of trail the attestations shall be reported to. Required
EOF
}


function die
{
    echo "Error: $1" >&2
    exit 1
}


function check_arguments
{
    while getopts "hb:p:f:t:" opt; do
        case $opt in
            h)
                print_help
                exit 1
                ;;
            b)
                BASE_COMMIT=${OPTARG}
                ;;
            p)
                PROPOSED_COMMIT=${OPTARG}
                ;;
            f)
                FLOW_NAME=${OPTARG}
                ;;
            t)
                TRAIL_NAME=${OPTARG}
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
    if [ -z "${PROPOSED_COMMIT}" ]; then
        die "option -p <proposed-commit> is required"
    fi
    if [ -z "${FLOW_NAME}" ]; then
        die "option -f <flow-name> is required"
    fi
    if [ -z "${TRAIL_NAME}" ]; then
        die "option -t <trail-name> is required"
    fi
}


function get_commit_and_pull_request
{
    local commit_sha=$1; shift
    local result_file=$1; shift

    pr_data=$(gh pr list --search "${commit_sha}" --state merged --json author,latestReviews,mergeCommit,mergedAt,url)
    commit_data=$(gh search commits --hash "${commit_sha}" --json author)

    local compliant="true"
    combined_data=$(jq -n --arg commitsha "$commit_sha" --argjson commit "$commit_data" --argjson pr "$pr_data" \
      '{commit_sha: $commitsha, commit: $commit[0], pull_request: $pr[0]}')

    # Check for missing latestReviews or if that list is empty
    latest_reviews=$(echo "${pr_data}" | jq '.[0].latestReviews')
    if [ "$latest_reviews" = "null" ]; then
        combined_data=$(echo "${combined_data}" | jq '. += {"reason_for_non_compliance": "no pull-request"}')
        compliant="false"
    elif [ -z "$latest_reviews" -o "$latest_reviews" = "[]" ]; then
        combined_data=$(echo "${combined_data}" | jq '. += {"reason_for_non_compliance": "no reviewers"}')
        compliant="false"
    else
        # Find the entry where 'state' is APPROVED
        pr_author=$(echo "${pr_data}" | jq '.[0].author.login')
        reviews_length=$(echo "${pr_data}" | jq '.[0].latestReviews | length')
        for i in $(seq 0 $(( reviews_length - 1 )))
        do
            review=$(echo "${pr_data}" | jq ".[0].latestReviews[$i]")
            state=$(echo "$review" | jq ".state")
            if [ "$state" = '"APPROVED"' ]; then
                break
            fi
        done
        if [ "$state" != '"APPROVED"' ]; then
            combined_data=$(echo "${combined_data}" | jq '. += {"reason_for_non_compliance": "no state:APPROVED review"}')
            compliant="false"
        else
            # Fail if latest reviewer and auther is the same person
            review_author=$(echo "$review" | jq ".author.login")
            if [ "${review_author}" = "${pr_author}" ]; then
                combined_data=$(echo "${combined_data}" | jq '. += {"reason_for_non_compliance": "committer and approver are the same person"}')
                compliant="false"
            fi
        fi
    fi

    # Make sure that true/false are not quoted
    combined_data=$(echo "${combined_data}" | jq ". += {\"compliant\": $compliant}")
    echo "${combined_data}" > ${result_file}

    if [ "${compliant}" == "true" ]; then
        return 0
    else
        return 1
    fi
}

function get_commit_and_pr_data_and_report_to_kosli
{
    local base_commit=$1; shift
    local proposed_commit=$1; shift
    local commit_pull_request_flow=$1; shift
    local trail_name=$1; shift

    commits=($(gh api repos/:owner/:repo/compare/${base_commit}...${proposed_commit} -q '.commits[].sha'))
    for commit_sha in "${commits[@]}"; do
        short_commit_sha=${commit_sha:0:7}
        local file_name="commit_pr_${short_commit_sha}.json"
        local compliant="true"
        get_commit_and_pull_request ${commit_sha} ${file_name} || compliant="false"
        kosli attest generic \
            --name=commit_${short_commit_sha} \
            --compliant=${compliant} \
            --attachments=${file_name} \
            --flow=${commit_pull_request_flow} \
            --trail=${trail_name}
        rm ${file_name}
    done
}


function main
{
    check_arguments "$@"
    get_commit_and_pr_data_and_report_to_kosli ${BASE_COMMIT} ${PROPOSED_COMMIT} ${FLOW_NAME} ${TRAIL_NAME}
}

main "$@"
