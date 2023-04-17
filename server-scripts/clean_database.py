#!/usr/bin/env python3

import documentdb
from lib import on_localhost


def clean_database(mongo):
    if on_localhost():
        client = documentdb.MONGO_CLIENT
        db_names = client.list_database_names()
        for db_name in db_names:
            if db_name not in ['admin', 'config', 'local']:
                client.drop_database(db_name)

        documentdb.wait_till_ready_or_raise(mongo, secs=10)


if __name__ == "__main__":
    from model import Externals
    clean_database(Externals().mongo)
