#!/usr/bin/env bash


SCRIPT_NAME=get_file_from_test_integration.sh
FROM_FILE_PATH=""
TO_FILE_NAME=""

die()
{
    echo "Error: $1" >&2
    exit 1
}

print_help()
{
    cat <<EOF
Usage: $SCRIPT_NAME [options]  <DOCKER_FILE> [LOCAL_FILE]

Get a file from test_integration

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
    
    if [ $# -lt 1 ]; then
        echo "Error: Missing DOCKER_FILE argument"
        exit 1
    fi
    FROM_FILE_PATH=$1; shift
    if [ $# -eq 1 ]; then
        TO_FILE_NAME=$1; shift
    else
        TO_FILE_NAME=$(basename "${FROM_FILE_PATH}")
    fi
}

main()
{
    check_arguments "$@"
    docker exec cli_kosli_server sh -c "cat ${FROM_FILE_PATH}" > ${TO_FILE_NAME} || die "Failed to get file"
}

main "$@"
