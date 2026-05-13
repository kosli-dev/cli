#!/usr/bin/env bash

KOSLI_ORG=kosli-public
KOSLI_FLOW=cli-release

#kosli list trails --flow $KOSLI_FLOW
kosli evaluate trails "a2ee562" --flow cli --policy four-eyes-policy.rego --params '{"attestation_name": "pr"}' --show-input --output json

#echo "Attesting release evaluation result to trail ${CURRENT_TAG}..."
#kosli attest custom \
#    --type "four-eyes-result" \
#    --name "four-eyes-result" \
#    --attestation-data "${EVAL_FILE}" \
#    --attachments "${SCRIPT_DIR}/four-eyes-policy.rego" \
#    --trail "${CURRENT_TAG}" \
#    --flow "${KOSLI_FLOW}"
