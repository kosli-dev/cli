import os
import re
import json
import subprocess
import boto3
import time
import sys


kosli_flow_name = os.environ.get("KOSLI_FLOW_NAME", "")
log_bucket_name = os.environ.get("LOG_BUCKET_NAME", "")
ecs_exec_event_wait_timeout = os.environ.get("ECS_EXEC_EVENT_WAIT_TIMEOUT", 300)

def get_ecs_exec_event_principal_id(ecs_exec_session_id):
    client = boto3.client('cloudtrail')
    principal_id = ''
    ecs_exec_event_wait_count = 0
    while principal_id == '' and ecs_exec_event_wait_count < ecs_exec_event_wait_timeout:
        time.sleep(10)
        ecs_exec_event_wait_count += 10
        response = client.lookup_events(
            LookupAttributes=[{'AttributeKey': 'EventName', 'AttributeValue': 'ExecuteCommand'}]
        )
        events = response.get('Events')
        for event in events:
            event_message = json.loads(event['CloudTrailEvent'])
            response_elements = event_message.get('responseElements')
            if response_elements and response_elements.get('session').get('sessionId') == ecs_exec_session_id:
                principal_id = event_message['userIdentity']['sessionContext']['sessionIssuer']['principalId']
                break
        print(f'Waiting 10 sec more for the ECS exec event with the sessionId={ecs_exec_session_id} to apear in the Cloudtrail...', file=sys.stderr)
    if principal_id == '':
        print(f'Something went wrong, can\'t get ECS exec event with the sessionId={ecs_exec_session_id}.', file=sys.stderr)
        return None
    else:
        print(f'Done. Current ECS exec event sessionId is {ecs_exec_session_id}.', file=sys.stderr)
        return principal_id

def lambda_handler(event, context):
    try:
        # Extract S3 object key from the event data
        s3_object_key = event['detail']['object']['key']

        # Get ECS session id by extracting it from the S3 object key
        ecs_exec_session_id = s3_object_key.split('/')[-1].split('.log')[0]
        print(f'ECS_EXEC_SESSION_ID is {ecs_exec_session_id}')

        # Download the log file from S3
        s3_client = boto3.client('s3')
        local_log_file_path = f'/tmp/{ecs_exec_session_id}.log'
        s3_client.download_file(log_bucket_name, s3_object_key, local_log_file_path)

        print(f'Getting principal id of the current ECS exec session...')
        ecs_exec_principal_id = get_ecs_exec_event_principal_id(ecs_exec_session_id)

        # Create Kosli trail (if it is already exists, it will just add a report event to the existing trail)
        print(f'Creating Kosli trail {ecs_exec_principal_id} within {kosli_flow_name} flow.', file=sys.stderr)

        kosli_client = subprocess.run(['./kosli', 'begin', 'trail', ecs_exec_principal_id,
                                                '--template-file=evidence-template.yml',
                                                f'--flow={kosli_flow_name}'])

        # Upload the log file to the Kosli
        print('Uploading ECS exec log file to the Kosli...', file=sys.stderr)
        subprocess.run(['./kosli', 'attest', 'generic', 
                         f'--flow={kosli_flow_name}', 
                         f'--trail={ecs_exec_principal_id}',
                         '--name=command-logs',
                         f'--evidence-paths={local_log_file_path}'])

        return {
            'statusCode': 200,
            'body': 'Handler executed successfully'
        }
    except Exception as e:
        print(f"An error occurred: {str(e)}", file=sys.stderr)
        return {
            'statusCode': 500,
            'body': 'Handler encountered an error'
        }
