#!/usr/bin/env bash
set -Eeu

SCRIPT_NAME="create_never_alone_trail.sh"
FLOW_NAME=""
TRAIL_NAME=""
START_COMMIT_SHA=""
END_COMMIT_SHA=""
SOURCE_FLOW_NAME=""
SOURCE_ATTESTATION_NAME=""
PARENT_FLOW_NAME=""
PARENT_TRAIL_NAME=""
KOSLI_HOST=${KOSLI_HOST:-https://app.kosli.com}


function print_help
{
    cat <<EOF
Use: $SCRIPT_NAME [options]

Script to create a trail for collecting never-alone information from multiple commits.
Collects all commits between start-commit-sha and end-commit-sha and use it as a template for the trail.

Options are:
  -h                           Print this help menu
  -f <flow-name>               Name of kosli flow to report each commits never-alone compliance. Required
  -t <trail-name>              Name of the trail to report each commits never-alone compliance. Required
  -b <start-commit-sha>        Start commit sha, used for creating list of commits. Required
  -c <end-commit-sha>          End commit sha, used for creating list of commits. Required
  -s <source-flow-name>        Name of kosli flow where never-alone-data for each commit is stored. Required
  -n <source-attestation-name> Attestation name used for never-alone-data for each commit. Required
  -p <parent-flow-name>        Send an attestation about the never-alone-trail to the parent-flow. Optional
  -q <parent-trail-name>       Trail name of parent flow where the report shall be sent. Optional
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
    while getopts "hf:t:b:c:s:n:p:q:" opt; do
        case $opt in
            h)
                print_help
                exit 1
                ;;
            f)
                FLOW_NAME=${OPTARG}
                ;;
            t)
                TRAIL_NAME=${OPTARG}
                ;;
            b)
                START_COMMIT_SHA=${OPTARG}
                ;;
            c)
                END_COMMIT_SHA=${OPTARG}
                ;;
            s)
                SOURCE_FLOW_NAME=${OPTARG}
                ;;
            n)
                SOURCE_ATTESTATION_NAME=${OPTARG}
                ;;
            p)
                PARENT_FLOW_NAME=${OPTARG}
                ;;
            q)
                PARENT_TRAIL_NAME=${OPTARG}
                ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2
                exit 1
                ;;
        esac
    done

    if [ -z "${FLOW_NAME}" ]; then
        die "option -f <flow-name> is required"
    fi
    if [ -z "${TRAIL_NAME}" ]; then
        die "option -t <trail-name> is required"
    fi
    if [ -z "${START_COMMIT_SHA}" ]; then
        die "option -b <start-commit-sha> is required"
    fi
    if [ -z "${END_COMMIT_SHA}" ]; then
        die "option -c <end-commit-sha> is required"
    fi
    if [ -z "${SOURCE_FLOW_NAME}" ]; then
        die "option -s <source-flow-name> is required"
    fi
    if [ -z "${SOURCE_ATTESTATION_NAME}" ]; then
        die "option -n <source-attestation-name> is required"
    fi
    if { [[ -n "$PARENT_FLOW_NAME" && -z "$PARENT_TRAIL_NAME" ]] || [[ -z "$PARENT_FLOW_NAME" && -n "$PARENT_TRAIL_NAME" ]]; }; then
        die "You must provide either both options -p <parent-flow-name> and -q <parent-trail-name>, or neither"
    fi
}

function begin_trail_with_template
{
    local -r flow_name=$1; shift
    local -r trail_name=$1; shift
    local -r commit_shas=("$@")
    local -r trail_template_file_name="review_trail.yaml"

    # Create a template yaml file with a trail level generic attestation for each commit
    {
    cat <<EOF
version: 1
trail:
  attestations:
EOF

    for commit_sha in "${commit_shas[@]}"; do
        echo "    - name: ${commit_sha}"
        echo "      type: generic"
    done
    } > ${trail_template_file_name}

    # Begin trail with this template
    kosli begin trail ${trail_name} \
        --flow=${flow_name} \
        --template-file=${trail_template_file_name}
}

function echo_never_alone_attestation_in_trail
{
    local -r source_flow_name=$1; shift
    local -r source_trail_name=$1; shift
    local -r source_attestation_name=$1; shift
    local -r never_alone_json_file_name=$(mktemp)

    local -r source_never_alone_attestation_url="${KOSLI_HOST}/api/v2/attestations/${KOSLI_ORG}/${source_flow_name}/trail/${source_trail_name}/${source_attestation_name}"
    http_code=$(curl -X 'GET' \
        --user ${KOSLI_API_TOKEN}:unused \
        "${source_never_alone_attestation_url}" \
        -H 'accept: application/json' \
        --output "${never_alone_json_file_name}" \
        --write-out "%{http_code}" \
        --silent)

    if [[ ${http_code} -lt 200 || ${http_code} -gt 299 ]] ; then
        # Error in curl command so print error and return empty array
        >&2 cat "${never_alone_json_file_name}"
        echo "[]"
        return
    fi

    cat "${never_alone_json_file_name}"
}

function set_never_alone_compliance
{
    local -r never_alone_data=$1; shift
    local pr_data compliant reviews pr_author reviews_length review state review_author
    
    COMPLIANT_STATUS="false"
    REASON_FOR_NON_COMPLIANT="Pull-request has not been approved by someone other than pr-author"
    pr_data=$(echo "${never_alone_data}" | jq '.user_data.pullRequest')
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

function attest_commit_trail_never_alone
{
    # Evaluate never-alone-data for this commit (from source commit trail) and attest compliance to this trail
    local -r flow_name=$1; shift
    local -r trail_name=$1; shift
    local -r commit_sha=$1; shift
    local -r source_flow_name=$1; shift
    local -r source_attestation_name=$1; shift
    
    local -r source_trail_name=${commit_sha:0:7}
    local url_to_source_attestation never_alone_data latest_never_alone_data compliant

    COMPLIANT_STATUS="false"
    never_alone_data=$(echo_never_alone_attestation_in_trail ${source_flow_name} ${source_trail_name} ${source_attestation_name})
    if [ "${never_alone_data}" != "[]" ]; then
        latest_never_alone_data=$(echo "${never_alone_data}" | jq '.[-1]')
        url_to_source_attestation=$(echo $latest_never_alone_data | jq -r '.html_url')
        set_never_alone_compliance "${latest_never_alone_data}"
        if [ "${COMPLIANT_STATUS}" == "true" ]; then
            kosli attest generic \
                --flow=${flow_name} \
                --trail=${trail_name} \
                --name="${commit_sha}" \
                --commit=${commit_sha} \
                --compliant="true" \
                --annotate="never_alone_data=${url_to_source_attestation}"
        else        
            kosli attest generic \
                --flow=${flow_name} \
                --trail=${trail_name} \
                --name="${commit_sha}" \
                --commit=${commit_sha} \
                --compliant="false" \
                --annotate="never_alone_data=${url_to_source_attestation}" \
                --annotate="reason_for_non_compliance=${REASON_FOR_NON_COMPLIANT}"
        fi
    fi
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

function main
{
    check_arguments "$@"
    # Use gh instead of git so we can keep the commit depth of 1. The order are from oldest commit to newest
    local -r commits=($(gh api repos/:owner/:repo/compare/${START_COMMIT_SHA}...${END_COMMIT_SHA} -q '.commits[].sha'))

    begin_trail_with_template ${FLOW_NAME} ${TRAIL_NAME} "${commits[@]}"
    
    local trail_compliance="true"
    for commit in "${commits[@]}"; do        
        attest_commit_trail_never_alone ${FLOW_NAME} ${TRAIL_NAME} ${commit} ${SOURCE_FLOW_NAME} ${SOURCE_ATTESTATION_NAME}
        if [ "${COMPLIANT_STATUS}" == "false" ]; then
            trail_compliance="false"
        fi
    done

    if [ -n "${PARENT_FLOW_NAME}" ]; then
        attest_never_alone_trail_to_parent  ${FLOW_NAME} ${TRAIL_NAME} ${PARENT_FLOW_NAME} ${PARENT_TRAIL_NAME} ${trail_compliance}
    fi
}

main "$@"
