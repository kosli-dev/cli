import boto3
import json
import subprocess
import sys
import os


kosli_flow_name = os.environ.get("KOSLI_FLOW_NAME", "")


def describe_ecs_task(cluster, task_arn):
    ecs_client = boto3.client('ecs')
    response = ecs_client.describe_tasks(cluster=cluster, tasks=[task_arn])
    return response

def lambda_handler(event, context):
    try:
        ecs_exec_session_id = event['detail']['responseElements']['session']['sessionId']
        print(f"ECS_EXEC_SESSION_ID is {ecs_exec_session_id}", file=sys.stderr)

        ecs_exec_principal_id = event['detail']['userIdentity']['sessionContext']['sessionIssuer']['principalId']

        # Create Kosli trail (if it is already exists, it will just add a report event to the existing trail)
        print(f'Creating Kosli trail {ecs_exec_principal_id} within {kosli_flow_name} flow.')
    
        kosli_client = subprocess.check_output(['./kosli', 'begin', 'trail', ecs_exec_principal_id,
                                                '--template-file=evidence-template.yml',
                                                f'--flow={kosli_flow_name}'])

        # Get and report ECS exec session user identity (ARN of the IAM role that initiated the session)
        ecs_exec_user_identity = event['detail']['userIdentity']['arn']
        with open('/tmp/user-identity.json', 'w') as user_identity_file:
            json.dump({"ecs_exec_role_arn": ecs_exec_user_identity}, user_identity_file)

        print("Reporting ECS exec user identity to the Kosli...", file=sys.stderr)
        subprocess.run(['./kosli', 'attest', 'generic',
                        f'--flow={kosli_flow_name}', 
                        f'--trail={ecs_exec_principal_id}',
                        '--name=user-identity',
                        '--evidence-paths=/tmp/user-identity.json',
                        '--user-data=/tmp/user-identity.json'])

        # Get and report ECS exec session service identity
        ecs_exec_task_arn = event['detail']['responseElements']['taskArn']
        ecs_exec_cluster = event['detail']['requestParameters']['cluster']
        ecs_task_info = describe_ecs_task(cluster=ecs_exec_cluster, task_arn=ecs_exec_task_arn)
        ecs_exec_task_group = ecs_task_info['tasks'][0]['group']

        with open('/tmp/service-identity.json', 'w') as service_identity_file:
            json.dump({"ecs_exec_service_identity": ecs_exec_task_group}, service_identity_file)

        print("Reporting ECS exec service identity to the Kosli...", file=sys.stderr)
        subprocess.run(['./kosli', 'attest', 'generic',
                        f'--flow={kosli_flow_name}', 
                        f'--trail={ecs_exec_principal_id}',
                        '--name=service-identity',
                        '--evidence-paths=/tmp/service-identity.json',
                        '--user-data=/tmp/service-identity.json'])

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
