#!/usr/bin/env bash
set -Eeu

SCRIPT_NAME=check_secret_expire.sh
ROOT_DIR=$(dirname $(readlink -f $0))/..
NOW_DATE=$(date +%Y-%m-%d)

print_help()
{
    cat <<EOF
Usage: $SCRIPT_NAME <options> [yyyy-mm-dd]

Will search all txt-files in secrets directory to see if any of them
has a secret that has expired. You can specify a date if you want to
know if something expires in the future

Options are:
  -h          Print this help menu
EOF
}

check_arguments()
{
    while getopts "h" opt; do
        case $opt in
            h)
                print_help
                exit 1
                ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2
                exit 1
                ;;
        esac
    done

    # Remove options from command line
    shift $((OPTIND-1))

    if [ $# -eq 1 ]; then
        NOW_DATE=$1; shift
    fi
}

echo_if_secret_expired()
{
    local file=$1; shift
    local now_date=$1; shift
    local expire_date now_sec expire_sec
    expire_date=$(grep "secret-expire:" ${file} | sed "s/secret-expire: *//")

    if [[ ! "${now_date}" < "${expire_date}" ]]; then
        grep "secret-name:" ${file} | sed "s/secret-name: */  /" | tr '\n' ' '
        grep "secret-expire:" ${file}
        return 1
    fi
    return 0
}

main()
{
    check_arguments "$@"
    local file
    local result=0
    echo "The following is a list of secrets in 'secrets/*txt' which will have expired on ${NOW_DATE}"
    for file in ${ROOT_DIR}/secrets/*txt; do
        echo_if_secret_expired ${file} ${NOW_DATE} || result=1
    done
    return $result
}

main "$@"
