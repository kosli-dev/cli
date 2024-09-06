#!/usr/bin/env bash
set -Eeu

SCRIPT_NAME="create_never_alone_trail.sh"
FLOW_NAME=""
TRAIL_NAME=""
BASE_COMMIT=""
CURRENT_COMMIT=""
SOURCE_FLOW=""
SOURCE_ATTESTATION_NAME=""
PARENT_FLOW=""
PARENT_TRAIL=""
KOSLI_HOST=${KOSLI_HOST:-https://app.kosli.com}


function print_help
{
    cat <<EOF
Use: $SCRIPT_NAME [options]

Script to create a trail for collecting never-alone information from multiple commits.
Collects all commits between base-commit and proposed-commit and use it as a template for the trail.

Options are:
  -h                    Print this help menu
  -f <flow-name>        Name of kosli flow to report combined never-alone info to. Required
  -t <trail-name>       Name of the trail that the reviews shall be reported to. Required
  -b <base-commit-sha>  Old commit sha, used as base for creating list of commits. Required
  -c <commit-sha>       Current commit sha, used as the end point for creating list of commits. Required
  -s <source-flow>      Name of kosli flow where the never-alone-data data are stored. Required
  -n <attestation-name> Attestation name used for never-alone-data. Required
  -p <parent-flow>      Send an attestation about the never-alone-trail to the parent-flow. Optional
  -q <parent-trail>     Trail name of parent flow where the report shall be sent. Optional
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
                BASE_COMMIT=${OPTARG}
                ;;
            c)
                CURRENT_COMMIT=${OPTARG}
                ;;
            s)
                SOURCE_FLOW=${OPTARG}
                ;;
            n)
                SOURCE_ATTESTATION_NAME=${OPTARG}
                ;;
            p)
                PARENT_FLOW=${OPTARG}
                ;;
            q)
                PARENT_TRAIL=${OPTARG}
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
    if [ -z "${BASE_COMMIT}" ]; then
        die "option -b <base-commit-sha> is required"
    fi
    if [ -z "${CURRENT_COMMIT}" ]; then
        die "option -c <commit-sha> is required"
    fi
    if [ -z "${SOURCE_FLOW}" ]; then
        die "option -s <source-flow> is required"
    fi
    if [ -z "${SOURCE_ATTESTATION_NAME}" ]; then
        die "option -n <attestation-name> is required"
    fi
    if { [[ -n "$PARENT_FLOW" && -z "$PARENT_TRAIL" ]] || [[ -z "$PARENT_FLOW" && -n "$PARENT_TRAIL" ]]; }; then
        die "You must provide either both options -p <parent-flow> and -q <parent-trail>, or neither"
    fi
}

function begin_trail_with_template
{
    local flow_name=$1; shift
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
        --flow=${flow_name} \
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
    local -r flow_name=$1; shift
    local -r trail_name=$1; shift
    local -r commit=$1; shift
    local -r source_flow=$1; shift
    local -r attestation_name=$1; shift
    local -r link_trail_name=${commit:0:7}

    local link_to_attestation never_alone_data latest_never_alone_data compliant

    link_to_attestation="${KOSLI_HOST}/${KOSLI_ORG}/flows/${source_flow}/trails/${link_trail_name}"
    never_alone_data=$(get_never_alone_attestation_in_trail ${source_flow} ${link_trail_name} ${attestation_name})
    if [ "${never_alone_data}" != "[]" ]; then
        latest_never_alone_data=$(echo "${never_alone_data}" | jq '.[-1]')
        compliant=$(get_never_alone_compliance "${latest_never_alone_data}")
        kosli attest generic \
            --flow ${flow_name} \
            --trail ${trail_name} \
            --commit ${commit} \
            --name="${commit}" \
            --compliant=${compliant} \
            --annotate never_alone_data="${link_to_attestation}"
    fi
}

function attest_never_alone_trail_to_parent
{
    local -r flow_name=$1; shift
    local -r trail_name=$1; shift
    local -r parent_flow=$1; shift
    local -r parent_trail=$1; shift

    never_alone_trail_link="${KOSLI_HOST}/${KOSLI_ORG}/flows/${flow_name}/trails/${trail_name}"
    kosli attest generic \
        --flow ${parent_flow} \
        --trail ${parent_trail} \
        --name never-alone-trail \
        --annotate never_alone_trail="${never_alone_trail_link}"
}

function main
{
    check_arguments "$@"
    # Use gh instead of git so we can keep the commit depth of 1. The order are from oldest
    # commit to newest
    commits=($(gh api repos/:owner/:repo/compare/${BASE_COMMIT}...${CURRENT_COMMIT} -q '.commits[].sha'))

    begin_trail_with_template ${FLOW_NAME} ${TRAIL_NAME} "${commits[@]}"
    
    for commit in "${commits[@]}"; do
        attest_commit_trail_never_alone ${FLOW_NAME} ${TRAIL_NAME} ${commit} ${SOURCE_FLOW} ${SOURCE_ATTESTATION_NAME}
    done

    if [ -n "${PARENT_FLOW}" ]; then
        attest_never_alone_trail_to_parent  ${FLOW_NAME} ${TRAIL_NAME} ${PARENT_FLOW} ${PARENT_TRAIL}
    fi
}

main "$@"
