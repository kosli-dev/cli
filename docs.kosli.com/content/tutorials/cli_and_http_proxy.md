---
title: "Using Kosli CLI with an HTTP proxy"
bookCollapseSection: false
weight: 511
---

# Using Kosli CLI with an HTTP proxy

In enterprises with strict network policies, you might want to communicate with Kosli via an HTTP proxy as the single point of egress communication with Kosli. 

This tutorial shows how you can setup an HTTP proxy and use it when communicating with Kosli via the CLI.

## TLDR

If you already have an HTTP proxy, [start using it with Kosli CLI](#use-the-http-proxy-with-kosli-cli)


{{<hint info>}}
In this tutorial, we will setup Tinyproxy (in docker) as an HTTP proxy on a Mac machine.
The same steps apply for different HTTP proxies and machines, but commands will differ.
{{</hint>}}


## Start the HTTP proxy

1. Start Tinyproxy using docker:

```shell {.command}
cat <<EOF > tinyproxy.conf
User nobody
Group nobody
Port 8888
EOF

docker run -p 8888:8888 -v $(PWD)/tinyproxy.conf:/etc/tinyproxy/tinyproxy.conf:ro kalaksi/tinyproxy
```



Now you have an HTTP proxy running at http://localhost:8888

## Use the HTTP proxy with Kosli CLI

To verify if the setup works, you can run this command to list environments of the public demo org `Cyber Dojo`:

```shell {.command}
kosli list envs --org cyber-dojo --http-proxy http://localhost:8888 --api-token <<your-token>>
```

Your request goes through the HTTP proxy and is then forwarded to Kosli. If successful, you should see a similar output to this:

```
NAME                         TYPE  LAST REPORT                LAST MODIFIED              TAGS
aws-beta                     ECS   2024-04-18T15:17:54+02:00  2024-04-18T15:17:54+02:00  [url=https://beta.cyber-dojo.org/]
aws-prod                     ECS   2024-04-18T15:17:57+02:00  2024-04-18T15:17:57+02:00  [url=https://cyber-dojo.org/]
terraform-state-differ-beta  S3    2024-04-18T15:18:23+02:00  2024-04-18T15:18:23+02:00  
terraform-state-differ-prod  S3    2024-04-18T15:18:17+02:00  2024-04-18T15:18:17+02:00 
```

All you need to do now is to use `--http-proxy http://localhost:8888` with your CLI commands.
Alternatively, you can add this to your kosli config so that you don't type it on each command:
`kosli config --http-proxy=http://localhost:8888`
