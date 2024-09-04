#!/usr/bin/env bash
set -Eeu

SCRIPT_NAME="create_review_trail_for_release.sh"
RELEASE_FLOW=""
TRAIL_NAME=""
BASE_COMMIT=""
PROPOSED_COMMIT=""
SOURCE_FLOW=""
SOURCE_ATTESTATION_NAME=""



function print_help
{
    cat <<EOF
Use: $SCRIPT_NAME [options]

Script to create a trail for a release. Collects all commits between base-commit
and proposed-commit and use it as a template for the trail.

Options are:
  -h                    Print this help menu
  -f <release-flow>     Name of kosli flow to report combined never-alone info to. Required
  -t <trail-name>       Name of the trail that the reviews shall be reported to. Required
  -b <base-commit>      Commit of previous release
  -p <proposed-commit>  Commit sha for release we are building now. Required
  -s <source-flow>      Name of kosli flow where the never-alone-data data are stored. Required
  -n <attestation-name> Attestation name used for never-alone-data. Required
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
    while getopts "hf:t:b:p:s:n:" opt; do
        case $opt in
            h)
                print_help
                exit 1
                ;;
            f)
                RELEASE_FLOW=${OPTARG}
                ;;
            t)
                TRAIL_NAME=${OPTARG}
                ;;
            b)
                BASE_COMMIT=${OPTARG}
                ;;
            p)
                PROPOSED_COMMIT=${OPTARG}
                ;;
            s)
                SOURCE_FLOW=${OPTARG}
                ;;
            n)
                SOURCE_ATTESTATION_NAME=${OPTARG}
                ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2
                exit 1
                ;;
        esac
    done

    if [ -z "${RELEASE_FLOW}" ]; then
        die "option -f <release-flow> is required"
    fi
    if [ -z "${TRAIL_NAME}" ]; then
        die "option -t <trail-name> is required"
    fi
    if [ -z "${BASE_COMMIT}" ]; then
        die "option -b <base-commit> is required"
    fi
    if [ -z "${PROPOSED_COMMIT}" ]; then
        die "option -p <proposed-commit> is required"
    fi
    if [ -z "${SOURCE_FLOW}" ]; then
        die "option -s <source-flow> is required"
    fi
    if [ -z "${SOURCE_ATTESTATION_NAME}" ]; then
        die "option -n <attestation-name> is required"
    fi
}

function begin_trail_with_template
{
    local release_flow=$1; shift
    local trail_name=$1; shift
    local commits=("$@")
    local trail_template_file_name="review_trail.yaml"

    # Create a template yaml file
    {
    cat <<EOF
version: 1
trail:
  attestations:
EOF

    for commit in "${commits[@]}"; do
        echo "    - name: ${commit}"
        echo "      type: generic"
    done
    } > ${trail_template_file_name}

    kosli begin trail ${trail_name} \
        --flow=${release_flow} \
        --description="$(git log -1 --pretty='%aN - %s')" \
        --template-file=${trail_template_file_name}
}

function get_never_alone_attestation_in_trail
{
    local -r source_flow=$1; shift
    local -r trail_name=$1; shift
    local -r attestation_name=$1; shift
    local -r curl_output_file=$(mktemp)

    http_code=$(curl -X 'GET' \
        --user ${KOSLI_API_TOKEN}:unused \
        "${KOSLI_HOST}/api/v2/attestations/${KOSLI_ORG}/${source_flow}/trail/${trail_name}/${attestation_name}" \
        -H 'accept: application/json' \
        --output "${curl_output_file}" \
        --write-out "%{http_code}" \
        --silent)

    if [[ ${http_code} -lt 200 || ${http_code} -gt 299 ]] ; then
        >&2 cat "${curl_output_file}"
        rm "${curl_output_file}"
        echo "[]"
        return
    fi

    cat "$curl_output_file"
    rm "${curl_output_file}"
}

function get_never_alone_compliance
{
    local -r never_alone_data=$1; shift
    local pr_data compliant reviews pr_author reviews_length review state review_author

    pr_data=$(echo "${never_alone_data}" | jq '.user_data.pullRequest')
    compliant="false"
    reviews=$(echo "${pr_data}" | jq '.reviews')
    pr_author=$(echo "${pr_data}" | jq '.author.login')
    reviews_length=$(echo "${pr_data}" | jq '.reviews | length')
    for i in $(seq 0 $(( reviews_length - 1 )))
    do
        review=$(echo "${pr_data}" | jq ".reviews[$i]")
        state=$(echo "$review" | jq ".state")
        review_author=$(echo "$review" | jq ".author.login")
        if [ "$state" = '"APPROVED"' -a "${review_author}" != "${pr_author}" ]; then
            compliant="true"
        fi
    done

    echo $compliant
}

function attest_commit_trail_never_alone
{
    local -r release_flow=$1; shift
    local -r trail_name=$1; shift
    local -r commit=$1; shift
    local -r source_flow=$1; shift
    local -r attestation_name=$1; shift
    local link_to_attestation never_alone_data latest_never_alone_data compliant

    link_to_attestation="${KOSLI_HOST}/${KOSLI_ORG}/flows/${source_flow}/trails/${commit}"
    never_alone_data=$(get_never_alone_attestation_in_trail ${source_flow} ${commit} ${attestation_name})
    if [ "${never_alone_data}" != "[]" ]; then
        latest_never_alone_data=$(echo "${never_alone_data}" | jq '.[-1]')
        compliant=$(get_never_alone_compliance "${latest_never_alone_data}")
        kosli attest generic \
            --flow ${release_flow} \
            --trail ${trail_name} \
            --name="${commit}" \
            --compliant=${compliant} \
            --external-url never-alone-data=${link_to_attestation}
    fi
}

function main
{
    check_arguments "$@"
    # Use gh instead of git so we can keep the commit depth of 1. The order are from oldest
    # commit to newest
    commits=($(gh api repos/:owner/:repo/compare/${BASE_COMMIT}...${PROPOSED_COMMIT} -q '.commits[].sha'))

    begin_trail_with_template ${RELEASE_FLOW} ${TRAIL_NAME} "${commits[@]}"
    
    for commit in "${commits[@]}"; do
        attest_commit_trail_never_alone ${RELEASE_FLOW} ${TRAIL_NAME} ${commit} ${SOURCE_FLOW} ${SOURCE_ATTESTATION_NAME}
    done
}

main "$@"
