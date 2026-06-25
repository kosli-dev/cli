#!/usr/bin/env python3

# Creates the standalone test users used by the Kosli CLI integration tests.
#
# This script is owned by the CLI repo (the test users are CLI test data). It is
# mounted into the server container at /app/test via docker-compose and executed
# there, so it relies on the server's `lib` and `model` packages being importable
# via PYTHONPATH=/app/src.

import hashlib

from lib import Sku
from model import Organizations, Users

# key == person-id, value == api-key
CLI_TEST_USERS = {
    "docs-cmd-test-user": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
    "acme-org": "v3OWZiYWu9G2IMQStYg9BcPQUQ88lJNNnTJTNq8jfvmkR1C5wVpHSs7F00JcB5i6OGeUzrKt3CwRq7ndcN4TTfMeo8ASVJ5NdHpZT7DkfRfiFvm8s7GbsIHh2PtiQJYs2UoN13T8DblV5C4oKb6-yWH73h67OhotPlKfVKazR-c",
    "iu-org": "qM9u2_grv6pJLbACwsMMMT5LIQy82tQj2k1zjZnlXti1smnFaGwCKW4jzk0La7ae9RrSYvEwCXSsXknD6YZqd-onLaaIUUKtEn6-B6yh53vWIe9EC5u85FCbKZjFbaicp_d0Me0Zcqq_KcCgrAZRX9xggl_pBb2oaCsNdllqNjk",
    "system-tests-user": "95-IeGBfyKdTteLdKidiAnXk6uMmV6jTkGM9v3DEtrQ",
}


def create_standalone_test_users(test_users):
    users = Users()
    orgs = Organizations()

    for user_name, api_key in test_users.items():
        uid = hashlib.sha256(user_name.encode("utf-8")).hexdigest()[0:24]
        login_data = {
            "userId": uid,
            "name": user_name,
            "email": "default@example.com",
            "picture": "",
        }
        users.create("descope", login_data)
        user = users.find_by_auth_user_id(login_data["userId"])
        user.completed_signup = True
        user.add_api_key(api_key, user, 0, "")
        user.auth_token = "213c18081df7f738ec479107b86f97ec678b1d54"

        orgs.create_shared(f"{user_name}-shared", sku=Sku().existing_orgs, owner=user)


if __name__ == "__main__":
    create_standalone_test_users(CLI_TEST_USERS)
