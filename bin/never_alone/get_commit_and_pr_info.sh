#!/usr/bin/env bash
set -Eeu

SCRIPT_NAME="get_commit_and_pr_info.sh"
COMMIT=""
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
                COMMIT=${OPTARG}
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

    if [ -z "${COMMIT}" ]; then
        die "option -c <commit-sha> is required"
    fi
    if [ -z "${NEVER_ALONE_JSON_FILENAME}" ]; then
        die "option -o <output-filename> is required"
    fi
}

function get_commit_data_using_graphql
{
    local -r commit=$1; shift
    gh api graphql \
    -F commit="${commit}" \
    -f query="$(gh repo view --json owner,name --jq '. | 
        "query($commit: String!) {
            repository(owner: \"\(.owner.login)\", name: \"\(.name)\") {
                object(expression: $commit) {
                    ... on Commit {
                        author {
                            name
                            email
                            date
                        }
                        committer {
                            name
                            email
                            date
                        }
                        message
                        tree {
                            oid
                        }
                        associatedPullRequests(first: 1) {
                            nodes {
                                url
                            }
                        }
                    }
                }
            }
        }"
    ')"
}


function get_never_alone_data
{
    local -r commit=$1; shift
    local -r result_file=$1; shift
    
    # We have seen that the 'gh search commits' sometimes return an empty list
    # Have added getting data with graphql also, and some echo messages further down
    # Only for debugging at the moment, but we could use graphql to get both commit and pr data
    commit_data_graphql=$(get_commit_data_using_graphql $commit)
    pr_data=$(gh pr list --search "${commit}" --state merged --json author,reviews,mergeCommit,mergedAt,reviewDecision,url)
    commit_data=$(gh search commits --hash "${commit}" --json commit)
    
    echo commit_data_graphql=$commit_data_graphql
    echo commit_data=$commit_data

    jq -n \
        --arg sha "$commit" \
        --argjson commit "$commit_data" \
        --argjson pullRequest "$pr_data" \
        '{
            sha: $sha,
            commit: $commit[0].commit,
            pullRequest: $pullRequest[0]
        }' > "${result_file}"
}


function main
{
    check_arguments "$@"
    get_never_alone_data ${COMMIT} ${NEVER_ALONE_JSON_FILENAME}
}


main "$@"
