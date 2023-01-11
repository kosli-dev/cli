---
title: "Step 1: Prerequisites and Kosli account"
bookCollapseSection: false
weight: 230
---

# Step 1: Prerequisites and Kosli account

To follow the tutorial, you will need to:

- Install both `Docker` and `docker-compose`.
- [Install the Kosli CLI](/getting_started/installation) and [set the `KOSLI_API_TOKEN` and `KOSLI_OWNER` environment variables](/getting_started/installation#getting-your-kosli-api-token).
- You can check your Kosli set up by running: 
    ```shell {.command}
    kosli pipeline ls
    ```
    which should return a list of pipelines or the message "No pipelines were found".

- Clone our quickstart-docker repository:
    ```shell {.command}
    git clone https://github.com/kosli-dev/quickstart-docker-example.git
    cd quickstart-docker-example
    ```

## Create Kosli account

You need a GitHub account to be able to use Kosli.  
Go to [app.kosli.com](https://app.kosli.com) and use "Sign up with GitHub" button to create a Kosli account. 
