#!/usr/bin/env bash
set -Eeu

SCRIPT_NAME="add_pr_to_release.sh"
SCRIPT_DIR=$(dirname $(readlink -f $0))

COMMIT_SHA=""
RELEASE_NAME=""
USER_DATA_FILENAME=never-alone-user-data.json
FLOW_NAME="cli-release-never-alone"
PARENT_FLOW_NAME="cli-release"
KOSLI_HOST=${KOSLI_HOST:-https://app.kosli.com}


function print_help
{
    cat <<EOF
Use: $SCRIPT_NAME [options]

Script to add PR evidence to an existing release. If a CI build is cancled the attest never-alone is not
reported to kosli. 

Options are:
  -h                    Print this help menu
  -r <release-name>     Name of the cli release (v2.10.17)
  -c <commit-sha>       Commit sha to report. Required
EOF
}


function die
{
    echo "Error: $1" >&2
    exit 1
}

function check_arguments
{
    while getopts "hc:r:" opt; do
        case $opt in
            h)
                print_help
                exit 1
                ;;
            c)
                COMMIT_SHA=${OPTARG}
                ;;
            r)
                RELEASE_NAME=${OPTARG}
                ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2
                exit 1
                ;;
        esac
    done

    if [ -z "${COMMIT_SHA}" ]; then
        die "option -c <commit-sha> is required"
    fi
    if [ -z "${RELEASE_NAME}" ]; then
        die "option -r <release-name> is required"
    fi
}

function set_never_alone_compliance
{
    local -r never_alone_data=$1; shift
    local pr_data reviews pr_author reviews_length review state review_author
    
    COMPLIANT_STATUS="false"
    REASON_FOR_NON_COMPLIANT="Pull-request has not been approved by someone other than pr-author"
    pr_data=$(echo "${never_alone_data}" | jq '.pullRequest')
    reviews=$(echo "${pr_data}" | jq '.reviews')
    pr_author=$(echo "${pr_data}" | jq '.author.login')
    reviews_length=$(echo "${pr_data}" | jq '.reviews | length')
    for i in $(seq 0 $(( reviews_length - 1 )))
    do
        review=$(echo "${pr_data}" | jq ".reviews[$i]")
        state=$(echo "$review" | jq ".state")
        review_author=$(echo "$review" | jq ".author.login")
        if [ "$state" == '"APPROVED"' -a "${review_author}" != "${pr_author}" ]; then
            COMPLIANT_STATUS="true"
            REASON_FOR_NON_COMPLIANT=""
        fi
    done
}

function attest_never_alone_trail_to_parent
{
    local -r flow_name=$1; shift
    local -r trail_name=$1; shift
    local -r parent_flow_name=$1; shift
    local -r parent_trail_name=$1; shift
    local -r trail_compliance=$1; shift

    never_alone_trail_url="${KOSLI_HOST}/${KOSLI_ORG}/flows/${flow_name}/trails/${trail_name}"
    kosli attest generic \
        --flow=${parent_flow_name} \
        --trail=${parent_trail_name} \
        --name=never-alone-trail \
        --compliant=${trail_compliance} \
        --annotate="never_alone_trail=${never_alone_trail_url}"
}

function attest_never_alone_data_to_release_never_alone
{
    local flow_name=$1; shift
    local release_name=$1; shift
    local commit_sha=$1; shift
    local full_commit_sha=$(git rev-parse ${commit_sha})

    ${SCRIPT_DIR}/get_commit_and_pr_info.sh -c ${commit_sha} -o ${USER_DATA_FILENAME} 

    never_alone_data=$(cat ${USER_DATA_FILENAME})

    PR_URL=$(echo ${never_alone_data} | jq -r '.pullRequest.url // empty')
    if [ -n "$PR_URL" ]; then
        pr_author_name=$(echo "${never_alone_data}" | jq -r '.pullRequest.author.name')
        review_decision=$(echo "${never_alone_data}" | jq -r '.pullRequest.reviewDecision')
        pr_url=$(echo "${never_alone_data}" | jq -r '.pullRequest.url')
        reviewers=$(echo "${never_alone_data}" | jq -r '.pullRequest.reviews[0].author.name')
        set_never_alone_compliance "${never_alone_data}"

        if [ "${COMPLIANT_STATUS}" == "true" ]; then
            kosli attest generic \
                --flow=${flow_name} \
                --trail=${release_name} \
                --name="${full_commit_sha}" \
                --commit=${full_commit_sha} \
                --compliant="true" \
                --annotate="pr_author_name=${pr_author_name}" \
                --annotate="review_decision=${review_decision}" \
                --annotate="pull_request=${pr_url}" \
                --annotate="reviewers=${reviewers}"
        else        
            kosli attest generic \
                --flow=${flow_name} \
                --trail=${release_name} \
                --name="${full_commit_sha}" \
                --commit=${full_commit_sha} \
                --compliant="false" \
                --annotate="pr_author_name=${pr_author_name}" \
                --annotate="review_decision=${review_decision}" \
                --annotate="pull_request=${pr_url}" \
                --annotate="reviewers=${reviewers}" \
                --annotate="reason_for_non_compliance=${REASON_FOR_NON_COMPLIANT}"
        fi

    else
        echo "No pull request found for ${commit_sha}"
    fi    
}

function get_trail_compliance
{
    local -r flow_name=$1; shift
    local -r trail_name=$1; shift
    kosli get trail ${trail_name} \
        --org=${KOSLI_ORG} \
        --flow=${flow_name} \
        --output json | jq -r ".compliance_status.is_compliant"

}

function main
{
    check_arguments "$@"
    attest_never_alone_data_to_release_never_alone ${FLOW_NAME} ${RELEASE_NAME} ${COMMIT_SHA}
    trail_compliance=$(get_trail_compliance ${FLOW_NAME} ${RELEASE_NAME})
    attest_never_alone_trail_to_parent  ${FLOW_NAME} ${RELEASE_NAME} ${PARENT_FLOW_NAME} ${RELEASE_NAME} ${trail_compliance}
}

main "$@"
