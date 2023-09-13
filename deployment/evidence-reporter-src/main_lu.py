import os
import re
import json
import subprocess
import boto3


kosli_audit_trail_name = os.environ.get("KOSLI_AUDIT_TRAIL_NAME", "")
kosli_step_name = os.environ.get("KOSLI_STEP_NAME", "")
log_bucket_name = os.environ.get("LOG_BUCKET_NAME", "")

def lambda_handler(event, context):
    # Extract S3 object key from the event data
    s3_object_key = event['detail']['object']['key']

    # Get ECS session id by extracting it from the S3 object key
    ecs_exec_session_id = s3_object_key.split('/')[-1].split('.log')[0]
    print(f'ECS_EXEC_SESSION_ID is {ecs_exec_session_id}')

    # Download the log file from S3
    s3_client = boto3.client('s3')
    local_log_file_path = f'/tmp/{ecs_exec_session_id}.log'
    s3_client.download_file(log_bucket_name, s3_object_key, local_log_file_path)

    # Check if the Kosli workflow already exists
    kosli_workflow_already_exists = False

    try:
        kosli_client = subprocess.check_output(['./kosli', 'list', 'workflows', 
                                                '--audit-trail', kosli_audit_trail_name, 
                                                '-o', 'json'])
        kosli_workflows_list = json.loads(kosli_client)
        for workflow in kosli_workflows_list:
            if workflow['id'] == ecs_exec_session_id:
                kosli_workflow_already_exists = True
                break
    except subprocess.CalledProcessError:
        pass

    if not kosli_workflow_already_exists:
        print(f'The Kosli workflow {ecs_exec_session_id} does not yet exist, creating it...')
        subprocess.call(['./kosli', 'report', 'workflow', 
                         '--audit-trail', kosli_audit_trail_name, 
                         '--id', ecs_exec_session_id])

    # Upload the log file to the Kosli
    print('Uploading ECS exec log file to the Kosli...')
    subprocess.call(['./kosli', 'report', 'evidence', 'workflow', '--audit-trail', kosli_audit_trail_name,
                     '-e', local_log_file_path, '--id', ecs_exec_session_id, '--step', kosli_step_name])
