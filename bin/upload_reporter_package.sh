#!/bin/bash -Eeu

KOSLI_CLI_VERSION="${TAG:1}"
curl -o merkely_$KOSLI_CLI_VERSION.tar.gz -L https://github.com/kosli-dev/cli/releases/download/v$KOSLI_CLI_VERSION/merkely_${KOSLI_CLI_VERSION}_linux_amd64.tar.gz
tar -xf merkely_$KOSLI_CLI_VERSION.tar.gz
chmod 755 deployment/reporter-lambda-src/*
zip -j kosli_lambda_$KOSLI_CLI_VERSION.zip deployment/reporter-lambda-src/* kosli
aws s3 cp kosli_lambda_$KOSLI_CLI_VERSION.zip s3://$S3_NAME/kosli_lambda_$KOSLI_CLI_VERSION.zip