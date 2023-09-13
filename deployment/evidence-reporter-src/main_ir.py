import boto3
import json
import subprocess
import sys
import os


kosli_audit_trail_name = os.environ.get("KOSLI_AUDIT_TRAIL_NAME", "")
kosli_step_name_user_identity = os.environ.get("KOSLI_STEP_NAME_USER_IDENTITY", "")
kosli_step_name_service_identity = os.environ.get("KOSLI_STEP_NAME_SERVICE_IDENTITY", "")

def describe_ecs_task(cluster, task_arn):
    ecs_client = boto3.client('ecs')
    response = ecs_client.describe_tasks(cluster=cluster, tasks=[task_arn])
    return response

def lambda_handler(event, context):
    try:
        ecs_exec_session_id = event['detail']['responseElements']['session']['sessionId']
        print(f"ECS_EXEC_SESSION_ID is {ecs_exec_session_id}", file=sys.stderr)

        # Check if workflow already exists. If not - create it.
        kosli_workflows_list = subprocess.run(['./kosli', 'list', 'workflows', 
                                               '--audit-trail', kosli_audit_trail_name, 
                                               '-o', 'json'], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        kosli_workflows_list = kosli_workflows_list.stdout.decode('utf-8')
        kosli_workflow_already_exists = ecs_exec_session_id in [workflow['id'] for workflow in json.loads(kosli_workflows_list)]
        
        if not kosli_workflow_already_exists:
            print(f"The Kosli workflow {ecs_exec_session_id} does not yet exist, creating it...", file=sys.stderr)
            subprocess.run(['./kosli', 'report', 'workflow', 
                            '--audit-trail', kosli_audit_trail_name, 
                            '--id', ecs_exec_session_id])

        # Get and report ECS exec session user identity (ARN of the IAM role that initiated the session)
        ecs_exec_user_identity = event['detail']['userIdentity']['arn']
        with open('/tmp/user-identity.json', 'w') as user_identity_file:
            json.dump({"ecs_exec_role_arn": ecs_exec_user_identity}, user_identity_file)

        print("Reporting ECS exec user identity to the Kosli...", file=sys.stderr)
        subprocess.run(['./kosli', 'report', 'evidence', 'workflow', 
                        '--audit-trail', kosli_audit_trail_name, 
                        '--user-data', '/tmp/user-identity.json', 
                        '--evidence-paths', '/tmp/user-identity.json', 
                        '--id', ecs_exec_session_id, 
                        '--step', kosli_step_name_user_identity])

        # Get and report ECS exec session service identity
        ecs_exec_task_arn = event['detail']['responseElements']['taskArn']
        ecs_exec_cluster = event['detail']['requestParameters']['cluster']
        ecs_task_info = describe_ecs_task(cluster=ecs_exec_cluster, task_arn=ecs_exec_task_arn)
        ecs_exec_task_group = ecs_task_info['tasks'][0]['group']

        with open('/tmp/service-identity.json', 'w') as service_identity_file:
            json.dump({"ecs_exec_service_identity": ecs_exec_task_group}, service_identity_file)

        print("Reporting ECS exec service identity to the Kosli...", file=sys.stderr)
        subprocess.run(['./kosli', 'report', 'evidence', 'workflow', 
                        '--audit-trail', kosli_audit_trail_name, 
                        '--user-data', '/tmp/service-identity.json', 
                        '--evidence-paths', '/tmp/service-identity.json', 
                        '--id', ecs_exec_session_id, 
                        '--step', kosli_step_name_service_identity])

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
