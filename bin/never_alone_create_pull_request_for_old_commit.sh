#!/usr/bin/env bash
set -Eeu

SCRIPT_NAME=never_alone_create_pull_request_for_old_commit.sh
BASE_COMMIT=""
MAIN_BRANCH="main"
TEMPORARY_FILE="never-alone-temporary-file.txt"


function print_help
{
    cat <<EOF
Usage: $SCRIPT_NAME [options] <git-commit>

Script to create a pull request on an old commit that is already on main/master

Options are:
  -h               Print this help menu
  -m <main-branch> Name of main/master branch. Default: ${MAIN_BRANCH}
EOF
}


function die
{
    echo "Error: $1" >&2
    exit 1
}


function check_arguments
{
    while getopts "hm:" opt; do
        case $opt in
            h)
                print_help
                exit 1
                ;;
            m)
                MAIN_BRANCH=${OPTARG}
                ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2
                exit 1
                ;;
        esac
    done

    # Remove options from command line
    shift $((OPTIND-1))

    if [ "$#" -eq 0 ]; then
        die "<git-commit> is a required parameter"
    fi

    local base_commit="$1"

    # Make sure we have a single long commit sha
    local full_base_commit=$(git rev-parse ${base_commit})
    if [ -z "$full_base_commit" ]; then
        die "Error: Unable to resolve commit SHA: $base_commit"
    fi

    BASE_COMMIT=${full_base_commit}
}


function create_pull_request_for_old_commit
{
    local base_commit=$1; shift
    local main_branch=$1; shift
    local temporary_file=$1; shift
    local branch="not-alone-pr-$base_commit"

    # Create a branch based on commit we want to approve and push a change to GitHub
    git checkout -b ${branch} ${base_commit}
    echo "Temporary file for ${base_commit}" > ${temporary_file}
    git add ${temporary_file}
    git commit -m "Added temporary_file"
    git push --set-upstream origin ${branch}

    # Create a pull request and auto merge it
    gh pr create --title "Review of old commit ${base_commit}" \
      --body "This PR is for reviewing the commit SHA ${base_commit}" \
      --head ${branch} \
      --base ${main_branch}
    gh pr merge --auto --squash --delete-branch || true

    # Remove the temporary file and push that
    rm ${temporary_file}
    git add ${temporary_file}
    git commit -m "Removed temporary_file"
    git push
    git checkout ${main_branch}
}


function main
{
    check_arguments "$@"
    create_pull_request_for_old_commit ${BASE_COMMIT} ${MAIN_BRANCH} ${TEMPORARY_FILE}
}

main "$@"
