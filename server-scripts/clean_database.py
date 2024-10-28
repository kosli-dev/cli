#!/usr/bin/env python3

import documentdb
from lib import on_localhost
from seams import external_mongo

CORE_DATABASE_NAMES = ['auth', 'environments', 'policies', 'projects']
INFRA_DATABASE_NAMES = ['admin', 'config', 'local']


def clean_database(mongo):
    print(f"Dropping databases ", end="")
    db_names = list(mongo.list_database_names())
    for db_name in db_names:
        if db_name not in INFRA_DATABASE_NAMES:
            print('.', end='')
            mongo.drop_database(db_name)
    print(f" {len(db_names)-3}", end='')
    for db_name in db_names:
        if db_name not in CORE_DATABASE_NAMES and db_name not in INFRA_DATABASE_NAMES:
            print(f" ({db_name[0:8]})", end='')
    print()


if __name__ == "__main__":
    if on_localhost():
        mongo = external_mongo()
        clean_database(mongo)
        documentdb.wait_till_ready_or_raise(mongo, secs=10)
