import os

db_config = {
    'user': os.environ['POSTGRES_USER'],
    'password': os.environ['POSTGRES_PASSWORD'],
    'host': os.environ['POSTGRES_HOST'],
    'db': os.environ['POSTGRES_DB'],
}


class Config(object):
    SQLALCHEMY_DATABASE_URI = "postgresql://{user}:{password}@{host}/{db}".format(**db_config)
    SQLALCHEMY_TRACK_MODIFICATIONS = False
