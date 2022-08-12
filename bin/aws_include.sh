# The POSSIBLE_AWS_SERVERS names must match the [profile <name>] in
# ~/.aws/config
POSSIBLE_AWS_SERVERS='staging|azure|prod|dnb|stacc|modulr'
AWS_VAULT_ENV_FILE_BASE=~/.aws/aws_vault_env

S3_PATH_STAGING="s3://merkely-temp"
S3_PATH_PROD="s3://merkely-prod-temp"
S3_PATH_DNB="s3://merkely-dnb-temp"
S3_PATH_STACC="s3://merkely-stacc-temp"
S3_PATH_MODULR="s3://merkely-modulr-temp"

## Disabled for now since Ewelinca has bash version 3.x.x
## Bash associative array
#declare -A S3_PATHS
#S3_PATHS[staging]="s3://merkely-temp"
##S3_PATHS[azure]=
#S3_PATHS[prod]="s3://merkely-prod-temp"
#S3_PATHS[dnb]="s3://merkely-dnb-temp"
#S3_PATHS[stacc]="s3://merkely-stacc-temp"
#S3_PATHS[modulr]="s3://merkely-modulr-temp"


# Notes about AWS tools.
# For commands that execute 'aws ecs' commands it is necessary
# to do a 'aws sso login' first.
# For commands that execute 'ecs-cli' commands it is necessary
# to do an additional 'aws-vault' command after the 'aws sso login'

die()
{
    echo "Error: $1" >&2
    exit 1
}

print_help_simple()
{
    cat <<EOF
Usage: ${SCRIPT_NAME} <${POSSIBLE_AWS_SERVERS}>

${HELP_STRING}

More help can be found here:
https://github.com/merkely-development/knowledge-base/blob/master/developer_settup.md#configure-awc-cli-to-create-new-profile

Options are:
  -h          Print this help menu
EOF
}

check_arguments_simple()
{
    while getopts "h" opt; do
        case $opt in
            h)
                print_help_simple
                exit 1
                ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2
                exit 1
                ;;
        esac
    done

    if [ $# -eq 0 ]; then
        die "Missing server. Must be one of '${POSSIBLE_AWS_SERVERS}'"
    fi

    export AWS_SERVER_NAME=$1
    if [[ ! "$AWS_SERVER_NAME" =~ ^($POSSIBLE_AWS_SERVERS)$ ]]; then
        die "Server must be one of '${POSSIBLE_AWS_SERVERS}'"
    fi
}

login_aws()
{
    local awsServerName=$1; shift
    if ! aws sts get-caller-identity --profile ${awsServerName} &> /dev/null; then
        aws sso login --profile ${awsServerName}
    fi
}

get_task_id()
{
    # Get all task arns
    local awsServerName=$1; shift
    local taskArns=$(aws ecs list-tasks \
      --cluster merkely \
      --output text \
      --profile ${awsServerName} \
      | sed "s/TASKARNS//")

    # Then get the newest one
    aws ecs describe-tasks \
      --cluster merkely \
      --tasks ${taskArns} \
      --profile ${awsServerName} \
      --query "tasks[] | reverse(sort_by(@, &createdAt)) | [].[createdAt,taskArn]" \
      --output text \
      | head -n 1 | sed 's#^.*/##'
}

is_vault_credentials_valid()
{
    local awsServerName=$1; shift
    local awsVaultEnvFile=${AWS_VAULT_ENV_FILE_BASE}_${awsServerName}

    # In the users home directory there is a file with the AWS_ variables needed to
    # use the ecs-cli tool. We check the expiration date to see if we can reuse the
    # variables.

    if [ -e ${awsVaultEnvFile} ]; then
        local expirationStr=$(grep AWS_SESSION_EXPIRATION ${awsVaultEnvFile} | sed "s/.*=//")
        if [ $(uname) = "Linux" ]; then
            local expirationTime=$(date --date=${expirationStr} +%s)
        else
            local expirationTime=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" ${expirationStr} +%s)
        fi
        local now=$(date +%s)
        if [ ${now} -lt ${expirationTime} ]; then
            source ${awsVaultEnvFile}
            return 0
        fi
    fi
    return 1
}

get_vault_credentials()
{
    local awsServerName=$1; shift
    local awsVaultEnvFile=${AWS_VAULT_ENV_FILE_BASE}_${awsServerName}

    # In the users home directory there is a file with the AWS_ variables needed to
    # use the ecs-cli tool. We check the expiration date to see if we can reuse the
    # variables.

    if is_vault_credentials_valid ${awsServerName}; then
        return 0
    fi

    make_vault_credentials ${awsServerName}
    source ${awsVaultEnvFile}
}

make_vault_credentials()
{
    local awsServerName=$1; shift
    local awsVaultEnvFile=${AWS_VAULT_ENV_FILE_BASE}_${awsServerName}

    json=$(aws-vault exec -j -d 12h ${awsServerName})
    local accessKeyId=$(echo $json | jq '.AccessKeyId')
    local secretAccessKey=$(echo $json | jq '.SecretAccessKey')
    local sessionToken=$(echo $json | jq '.SessionToken')
    local expiration=$(echo $json | jq '.Expiration' | sed 's/\"//g')
    local region=$(aws configure list --profile ${awsServerName} | grep region |  awk '{print $2}')

    cat << EOF > ${awsVaultEnvFile}
export AWS_REGION=${region}
export AWS_ACCESS_KEY_ID=${accessKeyId}
export AWS_SECRET_ACCESS_KEY=${secretAccessKey}
export AWS_SESSION_TOKEN=${sessionToken}
export AWS_SESSION_EXPIRATION=${expiration}
EOF
}

get_s3_path()
{
    local awsServerName=$1; shift
    case ${awsServerName} in
        staging)
            echo ${S3_PATH_STAGING}
            ;;
        prod)
            echo ${S3_PATH_PROD}
            ;;
        dnb)
            echo ${S3_PATH_DNB}
            ;;
        stacc)
            echo ${S3_PATH_STACC}
            ;;
        modulr)
            echo ${S3_PATH_MODULR}
            ;;
        *)
            return 1
            ;;
    esac
}

## Disabled for now since Ewelinca has bash version 3.x.x
#get_s3_path()
#{
#    local awsServerName=$1; shift
#    if [ ${S3_PATHS[$awsServerName]+_} ]; then
#        echo ${S3_PATHS[$awsServerName]}
#        return 0
#    else
#        return 1
#    fi
#}
