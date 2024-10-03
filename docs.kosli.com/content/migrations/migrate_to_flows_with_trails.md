---
title: "Migrate your flows reporting to use trails and attestations"
weight: 2
---

# Migrate your flows reporting to use trails and attestations

Initially, flows in Kosli represented artifacts and their evidence. [Trails was then introduced](https://www.kosli.com/blog/how-to-record-an-audit-trail-for-any-devops-process-with-kosli-trails/) to give users more flexibility to model business and/or software workflows they care about in Kosli flows. 

In October 2024, we will start migrating all flows data to use trails and all evidence to attestations. During (and after) the migration, deprecated CLI commands and API endpoints continue to work and get converted on-the-fly to use trails and attestations.

This guide aims to help users switch from the deprecated CLI commands to the newer first-class commands for flows with trails and attestations.

## CLI commands

Replace the commands in the first column with their counterpart from the middle column. Be sure to check each command documentation in the [CLI reference](https://docs.kosli.com/client_reference/) as it may have new or changed flags compared to the deprecated commands.

| Deprecated Command                        | Use this command instead              | Remarks                                                                                                    |
|-------------------------------------------|---------------------------------------|------------------------------------------------------------------------------------------------------------|
| kosli create flow --template ...          | kosli create flow --template-file ... | --template is deprecated. use --template-file instead.  Template files adhere to the format defined [here](https://docs.kosli.com/template_ref/).  Creating the flow can be skipped and Kosli will auto-create one for you when you report trails, artifacts or attestations on a flow name that does not exist. |
|           | kosli begin trail ... | Creating the trail can be skipped and Kosli will auto-create one for you when you report artifacts or attestations on a trail name that does not exist. |
| kosli report artifact ...                 | kosli attest artifact ...             |                                                                                                            |
| kosli report evidence artifact <type> ... | kosli attest <type> ...               |                                                                                                            |
| kosli report evidence commit <type> ...   | kosli attest <type> ...               |                                                                                                            |