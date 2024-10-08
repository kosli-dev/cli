#!/usr/bin/env bash
set -Eeu

REPO_NAME=$1; shift

export KOSLI_ORG=kosli
export KOSLI_FLOW="secrets-updated"

SECRETS_PATH="secrets/*.txt"
SECRETS_FILES_REGEXP="^secrets/.*\.txt"

get_soc_trail_name()
{
    local -r SOC_START_DAY=25
    local -r SOC_START_MONTH=2
    local current_day=$(date +%-d)
    local current_month=$(date +%-m)
    local current_year=$(date +%Y)

    if [[ ${current_month} -gt ${SOC_START_MONTH} || (${current_month} -eq ${SOC_START_MONTH} && ${current_day} -ge ${SOC_START_DAY}) ]]; then
        echo "soc-${current_year}-$((current_year + 1))"
    else
        echo "soc-$((current_year - 1))-${current_year}"
    fi
}

report_update_of_secrets_to_kosli()
{
    local repo_name=$1; shift
    local -r trail_name=$(get_soc_trail_name)
    local files_changed secret_name expire_date repository attestation_name

    files_changed=$(git diff --name-only HEAD^ HEAD ${SECRETS_PATH})
    for file in ${files_changed}; do
        secret_name=$(grep "^secret-name:" $file | sed "s/secret-name: *//")
        expire_date=$(grep "^secret-expire:" $file | sed "s/secret-expire: *//")
        secret_updated_by=$(grep "^secret-updated-by:" $file | sed "s/secret-updated-by: *//")
        attestation_name="${repo_name//\//_}-${secret_name//\//_}"

        kosli attest generic \
            --name=${attestation_name} \
            --annotate Secret_repository=${repo_name} \
            --annotate Secret_name=${secret_name} \
            --annotate Secret_expire="${expire_date}" \
            --annotate Secret_updated_by="${secret_updated_by}" \
            --trail=${trail_name}
    done

    files_deleted=$(git diff --name-only --diff-filter=D HEAD^ HEAD | grep ${SECRETS_FILES_REGEXP}) || true
    for file in ${files_deleted}; do
        secret_name=$(git show HEAD^:${file} | grep "^secret-name:" | sed "s/secret-name: *//")
        attestation_name="${repo_name//\//_}-${secret_name//\//_}"

        kosli attest generic \
            --name=${attestation_name} \
            --annotate Secret_repository=${repo_name} \
            --annotate Secret_name=${secret_name} \
            --annotate Secret_deleted="SECRET DELETED" \
            --trail=${trail_name}
    done

}

report_update_of_secrets_to_kosli ${REPO_NAME#*/}
