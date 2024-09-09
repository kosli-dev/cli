#!/usr/bin/env bash
set -Eeu

SCRIPT_NAME="get_commit_and_pr_info.sh"
COMMIT_SHA=""
NEVER_ALONE_JSON_FILENAME=""


function print_help
{
    cat <<EOF
Use: $SCRIPT_NAME [options]

Script to get commit and pull request info for a commit

Options are:
  -h                    Print this help menu
  -c <commit-sha>       Commit sha we are gathering data for. Required
  -o <output-filename>  Name of json file to save result: Required
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
    while getopts "hc:o:" opt; do
        case $opt in
            h)
                print_help
                exit 1
                ;;
            c)
                COMMIT_SHA=${OPTARG}
                ;;
            o)
                NEVER_ALONE_JSON_FILENAME=${OPTARG}
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
    if [ -z "${NEVER_ALONE_JSON_FILENAME}" ]; then
        die "option -o <output-filename> is required"
    fi
}

function get_commit_and_pr_data_using_graphql
{
    # This function is a replacement for two api calls
    # pr_data=$(gh pr list --search "${commit_sha}" --state merged --json author,reviews,mergeCommit,mergedAt,reviewDecision,url)
    # commit_data=$(gh search commits --hash "${commit_sha}" --json commit)
    # We have seen that the 'gh search commits' sometimes just return an empty list.
    # Combining them also reduces the number of API calls which is good since we have seen that we
    # have been rate limited
    local -r commit_sha=$1; shift

    repo_info=$(gh repo view --json owner,name --jq '. | {owner: .owner.login, name: .name}')
    repo_owner=$(echo "$repo_info" | jq -r '.owner')
    repo_name=$(echo "$repo_info" | jq -r '.name')

    gh api graphql \
        -F commit_sha="${commit_sha}" \
        -F owner="${repo_owner}" \
        -F name="${repo_name}" \
        -f query='
            query($commit_sha: String!, $owner: String!, $name: String!) {
                repository(owner: $owner, name: $name) {
                    object(expression: $commit_sha) {
                        ... on Commit {
                            author {
                                name
                                email
                                date
                            }
                            message
                            associatedPullRequests(first: 1) {
                                nodes {
                                    author {
                                        login
                                        ... on User {
                                            name
                                        }
                                    }
                                    mergeCommit {
                                        oid
                                    }
                                    mergedAt
                                    reviewDecision
                                    reviews(first: 100) {
                                        nodes {
                                            author {
                                                login
                                                ... on User {
                                                    name
                                                }
                                            }
                                            state
                                            submittedAt
                                            commit {
                                                oid
                                            }
                                        }
                                    }
                                    url
                                }
                            }                            
                        }
                    }
                }
            }'
}

function get_pr_data_using_graphql
{
    local -r commit=$1; shift


}

function get_never_alone_data
{
    local -r commit_sha=$1; shift
    local -r result_file=$1; shift
    
    commit_and_pr_data=$(get_commit_and_pr_data_using_graphql $commit_sha)
    commit_data=$(echo "${commit_and_pr_data}" | jq '{
        author: .data.repository.object.author,
        message: .data.repository.object.message
    }')
    pr_data=$(echo "${commit_and_pr_data}" | jq '.data.repository.object.associatedPullRequests.nodes[0] | .reviews = .reviews.nodes')

    jq -n \
        --arg commit_sha "$commit_sha" \
        --argjson commit_data "$commit_data" \
        --argjson pr_data "$pr_data" \
        '{
            commit_sha: $commit_sha,
            commit: $commit_data,
            pullRequest: $pr_data
        }' > "${result_file}"
}


function main
{
    check_arguments "$@"
    get_never_alone_data ${COMMIT_SHA} ${NEVER_ALONE_JSON_FILENAME}
}


main "$@"
