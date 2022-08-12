#!/bin/bash -Eeu

SCRIPT_NAME=docker_login_aws.sh
HELP_STRING="Does a docker login to AWS so we can fetch repositories from it"
SCRIPT_DIR=$(dirname $(readlink -f $0))
source ${SCRIPT_DIR}/aws_include.sh

get_aws_account_id()
{
    local awsServerName=$1; shift
    egrep "profile|sso_account_id" ~/.aws/config \
        | grep -A1 ${awsServerName} \
        | tail -1 \
        | sed "s/sso_account_id *= *//"
}

docker_login()
{
    local awsServerName=$1; shift
    local awsAccountId=$(get_aws_account_id ${awsServerName})
    aws ecr get-login-password --region eu-central-1 \
        | docker login --username AWS --password-stdin \
            ${awsAccountId}.dkr.ecr.eu-central-1.amazonaws.com
}


main()
{
    check_arguments_simple "$@"
    # On CI we don't need to log in. Use this variable to check that we are running on CI
    # and then return early
    if [ "${GITHUB_RUN_NUMBER:=""}" != "" ]; then
        exit 0
    fi
    login_aws ${AWS_SERVER_NAME} || die "Failed to do 'aws sso login' for '${AWS_SERVER_NAME}'"
    if ! is_vault_credentials_valid ${AWS_SERVER_NAME}; then
        get_vault_credentials ${AWS_SERVER_NAME} || die "Failed to create vault credentials for '${AWS_SERVER_NAME}'"
        docker_login ${AWS_SERVER_NAME} || die "Failed to do docker login to '${AWS_SERVER_NAME}'"
    fi
}

main "$@"
