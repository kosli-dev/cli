---
title: "v0.1.x > 2.0.0 migration"
bookCollapseSection: false
weight: 290
---
# v0.1.x > 2.0.0 migration

If you decided to migrate Kosli cli from version v0.1.x to v2.0.0 or later the table below can help you with figuring out how the commands have changed.  

{{< hint info >}}

Keep in mind that for some commands the [flag names or argument types](#flagsarguments) are also updated, so have a look at documentation for each command before switching.  
Reach out to us using [Slack](https://www.kosli.com/community/) if you find yourself in trouble.

{{< /hint >}}

## Commands

| v0.1.x                                                        | v2.0.0                                               |
|---------------------------------------------------------------|------------------------------------------------------|
| kosli approval get                                            | [kosli get approval](https://docs.kosli.com/client_reference/kosli_get_approval/)                                   |
| kosli approval ls                                             | [kosli list approvals](https://docs.kosli.com/client_reference/kosli_list_approvals/)                                   |
| kosli artifact get                                            | [kosli get artifact](https://docs.kosli.com/client_reference/kosli_get_artifact/)                                   |
| kosli artifact ls                                             | [kosli list artifacts](https://docs.kosli.com/client_reference/kosli_list_artifacts/)                                   |
| kosli assert artifact                                         | [kosli assert artifact](https://docs.kosli.com/client_reference/kosli_assert_artifact/)                                |
| kosli assert bitbucket-pullrequest                            | [kosli assert pullrequest bitbucket](https://docs.kosli.com/client_reference/kosli_assert_pullrequest_bitbucket/)                   |
| kosli assert environment                                      | [kosli assert snapshot](https://docs.kosli.com/client_reference/kosli_assert_snapshot/)                                |
| kosli assert github-pullrequest                               | [kosli assert pullrequest github](https://docs.kosli.com/client_reference/kosli_assert_pullrequest_github/)                      |
| kosli assert gitlab-mergerequest                              | [kosli assert pullrequest gitlab](https://docs.kosli.com/client_reference/kosli_assert_pullrequest_gitlab/)         |
| kosli assert status                                           | [kosli assert status](https://docs.kosli.com/client_reference/kosli_assert_status/)                                  |
| kosli commit report evidence bitbucket-pullrequest            | [kosli report evidence commit pullrequest bitbucket](https://docs.kosli.com/client_reference/kosli_report_evidence_commit_pullrequest_bitbucket/)                 |
| kosli commit report evidence generic                          | [kosli report evidence commit generic](https://docs.kosli.com/client_reference/kosli_report_evidence_commit_generic/)                    |
| kosli commit report evidence github-pullrequest               | [kosli report evidence commit pullrequest github](https://docs.kosli.com/client_reference/kosli_report_evidence_commit_pullrequest_github/)                   |
| kosli commit report evidence gitlab-mergerequest              | [kosli report evidence commit pullrequest gitlab](https://docs.kosli.com/client_reference/kosli_report_evidence_commit_pullrequest_gitlab/)                   |
| kosli commit report evidence junit                            | [kosli report evidence commit junit](https://docs.kosli.com/client_reference/kosli_report_evidence_commit_junit/)                    |
| kosli commit report evidence snyk                             | [kosli report evidence commit snyk](https://docs.kosli.com/client_reference/kosli_report_evidence_commit_snyk/)                    |
| kosli completion                                              | [kosli completion](https://docs.kosli.com/client_reference/kosli_completion/)                                     |
| kosli deployment get                                          | [kosli get deployment](https://docs.kosli.com/client_reference/kosli_get_deployment/)                                 |
| kosli deployment ls                                           | [kosli list deployments](https://docs.kosli.com/client_reference/kosli_list_deployments/)                                 |
| kosli environment allowedartifacts add                        | [kosli allow artifact](https://docs.kosli.com/client_reference/kosli_allow_artifact/)                                 |
| kosli environment declare                                     | [kosli create environment](https://docs.kosli.com/client_reference/kosli_create_environment/)                             |
| kosli environment diff                                        | [kosli diff snapshots](https://docs.kosli.com/client_reference/kosli_diff_snapshots/)                                 |
| kosli environment get                                         | [kosli get snapshot](https://docs.kosli.com/client_reference/kosli_get_snapshot/)                                   |
| kosli environment inspect                                     | [kosli get environment](https://docs.kosli.com/client_reference/kosli_get_environment/)                                |
| kosli environment log                                         | [kosli list snapshots](https://docs.kosli.com/client_reference/kosli_list_snapshots/)                                   |
| kosli environment log --long                                  | [kosli log environment](https://docs.kosli.com/client_reference/kosli_log_environment/)                               |
| kosli environment ls                                          | [kosli list environments](https://docs.kosli.com/client_reference/kosli_list_environments/)                                |
| kosli environment rename                                      | [kosli rename environment](https://docs.kosli.com/client_reference/kosli_rename_environment/)                             |
| kosli environment report docker                               | [kosli snapshot docker](https://docs.kosli.com/client_reference/kosli_snapshot_docker/)                                |
| kosli environment report ecs                                  | [kosli snapshot ecs](https://docs.kosli.com/client_reference/kosli_snapshot_ecs/)                                   |
| kosli environment report k8s                                  | [kosli snapshot k8s](https://docs.kosli.com/client_reference/kosli_snapshot_k8s/)                                   |
| kosli environment report lambda                               | [kosli snapshot lambda](https://docs.kosli.com/client_reference/kosli_snapshot_lambda/)                                |
| kosli environment report s3                                   | [kosli snapshot s3](https://docs.kosli.com/client_reference/kosli_snapshot_s3/)                                    |
| kosli environment report server                               | [kosli snapshot server](https://docs.kosli.com/client_reference/kosli_snapshot_server/)                                |
| kosli expect deployment                                       | [kosli expect deployment](https://docs.kosli.com/client_reference/kosli_expect_deployment/)                              |
| kosli pipeline deployment report                              | [kosli expect deployment](https://docs.kosli.com/client_reference/kosli_expect_deployment/)                              |
| kosli fingerprint                                             | [kosli fingerprint](https://docs.kosli.com/client_reference/kosli_fingerprint/)                                    |
| kosli pipeline approval assert                                | [kosli assert approval](https://docs.kosli.com/client_reference/kosli_assert_approval/)                                |
| kosli pipeline approval report                                | [kosli report approval](https://docs.kosli.com/client_reference/kosli_report_approval/)                                |
| kosli pipeline approval request                               | [kosli request approval](https://docs.kosli.com/client_reference/kosli_request_approval/)                               |
| kosli pipeline artifact report creation                       | [kosli report artifact](https://docs.kosli.com/client_reference/kosli_report_artifact/)                                |
| kosli pipeline artifact report evidence bitbucket-pullrequest | [kosli report evidence artifact pullrequest bitbucket](https://docs.kosli.com/client_reference/kosli_report_evidence_artifact_pullrequest_bitbucket/) |
| kosli pipeline artifact report evidence generic               | [kosli report evidence artifact generic](https://docs.kosli.com/client_reference/kosli_report_evidence_artifact_generic/)               |
| kosli pipeline artifact report evidence github-pullrequest    | [kosli report evidence artifact pullrequest github](https://docs.kosli.com/client_reference/kosli_report_evidence_artifact_pullrequest_github/)    |
| kosli pipeline artifact report evidence gitlab-mergerequest   | [kosli report evidence artifact pullrequest gitlab](https://docs.kosli.com/client_reference/kosli_report_evidence_artifact_pullrequest_gitlab/)    |
| kosli pipeline artifact report evidence junit                 | [kosli report evidence artifact junit](https://docs.kosli.com/client_reference/kosli_report_evidence_artifact_junit/)                 |
| pipeline artifact report evidence test                        | [kosli report evidence artifact junit](https://docs.kosli.com/client_reference/kosli_report_evidence_artifact_junit/)                 |
| kosli pipeline artifact report evidence snyk                  | [kosli report evidence artifact snyk](https://docs.kosli.com/client_reference/kosli_report_evidence_artifact_snyk/)                  |
| kosli pipeline declare                                        | [kosli create flow](https://docs.kosli.com/client_reference/kosli_create_flow/)                                    |
| kosli pipeline inspect                                        | [kosli get flow](https://docs.kosli.com/client_reference/kosli_get_flow/)                                       |
| kosli pipeline ls                                             | [kosli list flows](https://docs.kosli.com/client_reference/kosli_list_flows/)                                       |
| kosli search                                                  | [kosli search](https://docs.kosli.com/client_reference/kosli_search/)                                         |
| kosli status                                                  | [kosli status](https://docs.kosli.com/client_reference/kosli_status/)                                         |
| kosli version                                                 | [kosli version](https://docs.kosli.com/client_reference/kosli_version/)                                        |

## Flags/Arguments

| v0.1.x                                                        | v2.0.0                                               |
|---------------------------------------------------------------|------------------------------------------------------|
| Pipeline as argument (for some commands)               | **--flow**                              |
|  **--owner**                                                   | **--org**           |
|  **--sha256**                                                   | **--fingerprint**           |
|  **--pipeline**                                                   | **--flow**           |
|  **--pipelines**                                                   | **--flows**           |
|  **--evidence-type**                                                   | **--name**           |