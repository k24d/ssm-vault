# SSM Parameter Store from Docker

This document describes how to use SSM Vault from Docker and docker-compose.

## Docker and environment variables

Suppose you are developing a web application using two containers: "db" and "app".  Your `docker-compose.yml` might look like this:

```yaml
version: '3'
services:
  db:
    image: postgres:latest
    environment:
      - POSTGRES_USER=dbuser
      - POSTGRES_PASSWORD=my-secret-password
      - POSTGRES_DB=app

  app:
    build: .
    volumes:
      - .:/app
    environment:
      - POSTGRES_USER=dbuser
      - POSTGRES_PASSWORD=my-secret-password
      - POSTGRES_HOST=db
      - POSTGRES_DB=app
```

According to [the PostgreSQL Docker image](https://hub.docker.com/_/postgres), we can initialize a DB user and password by settings the environment variables `POSTGRES_USER` and `POSTGRES_PASSWORD`.  Since we use the same user and password for db and app, we define identical environment variables for both containers as above.

In our application side, the configuration file might look like this:

```python
# config.py

db_config = {
    'user': os.environ['POSTGRES_USER'],
    'password': os.environ['POSTGRES_PASSWORD'],
    'host': os.environ['POSTGRES_HOST'],
    'db': os.environ['POSTGRES_DB'],
}

SQLALCHEMY_DATABASE_URI = "postgresql://{user}:{password}@{host}/{db}".format(**db_config)
```

(See [examples/docker/config.py](examples/docker/config.py) for full example.)

### Using .env for environment variables

In this particular example, we don't need to protect our database password because there is no real data.  However, sometimes we have some secrets that cannot to be stored in the repository.  So, let's create a `.env` file, which is loaded by docker-compose automatically:

```bash
# .env
POSTGRES_PASSWORD=my-secret-password
```

Now we can modify our `docker-compose.yml` as follows:

```yaml
version: '3'
services:
  db:
    image: postgres:latest
    environment:
      - POSTGRES_USER=dbuser
      - POSTGRES_PASSWORD     # defined in .env
      - POSTGRES_DB=app

  app:
    build: .
    volumes:
      - .:/app
    environment:
      - POSTGRES_USER=dbuser
      - POSTGRES_PASSWORD     # defined in .env
      - POSTGRES_HOST=db
      - POSTGRES_DB=app
```

Good!  Now we can safely share the code by excluding `.env` from the repository.

### Using ssm-vault for environment variables

Let's replace `.env` by `ssm-vault`.  We can store any secrets in addition to application parameters in SSM Parameter Store:

```bash
$ ssm-vault write /app/dev/postgres/user -s
Enter text: dbuser

$ ssm-vault write /app/dev/postgres/password
Enter secret: ********
```

These values can be exposed as environment variables by `ssm-vault exec`:

```bash
$ ssm-vault exec -p /app/dev -- env | grep POSTGRES
POSTGRES_PASSWORD=my-secret-password
POSTGRES_USER=dbuser
```

Instead of using `.env`, we can start a container as follows:

```
$ ssm-vault exec -p /app/dev -- docker-compose run app
```

In general, `.env` is more suitable for personal access keys of each developer, while SSM Parameter Store is better for shared credentials like 3rd-party service's API keys.

### Using ssm-vault from the container

It's tedious to enter `ssm-vault exec` every time we start containers.  Let's embed `ssm-vault` in a container and use it when the container starts.

You can add the following lines to your `Dockerfile` in order to install `ssm-vault`:

```bash
# Install ssm-vault
ENV SSM_VAULT_VERSION v1.0.0
ENV SSM_VAULT_CHECKSUM 1f8cc1479cb5e2688eca81de4ba1ee99e4bc08a1c753b38b648a5a3bbbf4c474
ADD https://github.com/k24d/ssm-vault/releases/download/$SSM_VAULT_VERSION/ssm-vault-linux-amd64 /usr/local/bin/ssm-vault
RUN echo "$SSM_VAULT_CHECKSUM /usr/local/bin/ssm-vault" | sha256sum -c && chmod 755 /usr/local/bin/ssm-vault
```

(See [examples/docker/Dockerfile](examples/docker/Dockerfile) for full example.)

Then, run `ssm-vault exec` either by `ENTRYPOINT` or `CMD`:

```bash
# Run "/usr/local/bin/ssm-vault exec" as the entry point
ENTRYPOINT ["/usr/local/bin/ssm-vault", "exec", "-p", "/app/dev", "--"]

or

# Run "/usr/local/bin/ssm-vault exec" before the actual command
CMD ["/usr/local/bin/ssm-vault", "exec", "-p", "/app/dev", "--", "flask", "run"]
```

Before starting a container, you need to set up AWS access keys in order to run `ssm-vault` successfully.  I recommend that you start a local "metadata server" by running [aws-vault](https://github.com/99designs/aws-vault) in server mode.  This way you can forget about AWS access keys in your container because a session token is provided by the metedata server in the same way as Amazon EC2 or ECS:

```bash
# Start a metadata server (only once)
$ aws-vault exec development --server
...

# Start a container in a different terminal
$ docker-compose run app
```

In this case, you start a metadata server at the beginning of day, and your container will start loading secrets from SSM Parameter Store whenever you start a container.

:warning: Don't forget to set the environment variable `AWS_REGION` for your container.  A metadata server does not provide region information, so you need to specify one explicitly:

```
  app:
    environment:
      - AWS_REGION=ap-northeast-1
```

## Generating configuration files

It is often said that environment variables are less secure than configuration files.  You can create configuration files in Docker containers by using `ssm-vault render`.

Let's execute a "run script" (`run.sh`) in the container:

```bash
# Dockerfile

COPY ./run.sh /app/runs.sh
CMD ["/app/run.sh"]
```

Within the script, we can generate a configuration file as follows:

```bash
#!/bin/bash

/usr/local/bin/ssm-vault render /app/config.template -p /app/dev -o /app/config.py

export FLASK_APP=/app/app.py
exec flask run
```

In the config template file, we can embed parameter values:

```python
# config.template

db_config = {
    'user': os.environ.get('POSTGRES_USER', '{{aws_ssm_parameter "postgres/user"}}'),
    'password': os.environ.get('POSTGRES_PASSWORD', '{{aws_ssm_parameter "postgres/password"}}'),
    'host': os.environ.get('POSTGRES_HOST', 'db'),
    'db': os.environ.get('POSTGRES_DB', 'app'),
}
```

### Creating secret files

If you have multiline secret files, such as X.509 private keys, you can create them in your run script:

```bash
#!/bin/bash

/usr/local/bin/ssm-vault read /app/dev/private_key -o /app/private_key.pem -m 0600
```

### Multiple environments

You might have multiple application environments:

```
% ssm-vault tree
.
└── /app/
    ├── dev/
    │   ├── postgres/
    │   │   ├── password🔐 (alias/aws/ssm)
    │   │   └── user
    ├── production/
    │   └── postgres/
    │       ├── password🔐 (alias/aws/ssm)
    │       └── user
    └── staging/
        └── postgres/
            ├── password🔐 (alias/aws/ssm)
            └── user
```

In this case, you need to select a parameter path depending on the environment:

```bash
#!/bin/bash

PARAMETER_PATH="/app/${APP_ENV:-dev}"

/usr/local/bin/ssm-vault render /app/config.template -p $PARAMETER_PATH -o /app/config.py

export FLASK_APP=/app/app.py
exec flask run
```

Then, switch environments by setting an environment name:

```
$ docker-compose run -e APP_ENV=staging app
```
