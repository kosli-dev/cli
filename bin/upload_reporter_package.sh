#!/bin/bash -Eeu

KOSLI_CLI_VERSION=$TAG
make build
chmod 755 deployment/reporter-lambda-src/*
zip -j kosli_lambda_$KOSLI_CLI_VERSION.zip deployment/reporter-lambda-src/* kosli
aws s3 cp kosli_lambda_$KOSLI_CLI_VERSION.zip s3://$S3_NAME/kosli_lambda_$KOSLI_CLI_VERSION.zip