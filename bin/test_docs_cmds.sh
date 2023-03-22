#!/usr/bin/env bash
set -e

MD_FILE=$1
TEST_FILE=/tmp/test_docs_cmd.sh

[ -f ${MD_FILE} ] || exit 9

# Create test file
echo "#!/usr/bin/env bash" > ${TEST_FILE}
echo "set -e" >> ${TEST_FILE}
echo "set -x" >> ${TEST_FILE}

# Clean up for the simulating_a_devops_system test
echo "rm -rf /tmp/try-kosli" >> ${TEST_FILE}

# Export cli commands
echo "export KOSLI_HOST=http://localhost:8001"  >> ${TEST_FILE}
echo "export KOSLI_ORG=docs-cmd-test-user"  >> ${TEST_FILE}
echo "export KOSLI_API_TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY"  >> ${TEST_FILE}
echo ""  >> ${TEST_FILE}

# Extract all shell commands from md file and append to test file
sed -n '/^```.*command/,/^```/ p' < ${MD_FILE} | sed '/^```/ d;/put your kosli/ d'  >> ${TEST_FILE}

# Chmod and execute the test file
chmod 755 ${TEST_FILE}
${TEST_FILE}
RESULT=$?
echo
echo "Test of '${MD_FILE}' passed"
exit ${RESULT}
