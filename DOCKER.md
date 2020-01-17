# Docker secret management

This document describes how to use SSM Value with Docker and docker-compose during development.

## Configuration by environment variables

Suppose you are developing a web application using two containers: "db" and "app".

You can create `docker-compose.yml` as follows:

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

According to [the PostgreSQL Docker image](https://hub.docker.com/_/postgres), we can initialize a DB user and password by settings environment variables `POSTGRES_USER` and `POSTGRES_PASSWORD`.  Since we use the same user and password in db and app, we define identical environment variables for these containers as above.

In our application side, the configuration file might look like this:
(See [examples/docker/config.py](examples/docker/config.py) for full example.)

```python
db_config = {
    'user': os.environ['POSTGRES_USER'],
    'password': os.environ['POSTGRES_PASSWORD'],
    'host': os.environ['POSTGRES_HOST'],
    'db': os.environ['POSTGRES_DB'],
}

SQLALCHEMY_DATABASE_URI = "postgresql://{user}:{password}@{host}/{db}".format(**db_config)
```

### Using .env for environment variables

In this particular example, we don't need to protect our database password, but sometimes we have secrets that we don't want to store in our source code, including `docker-compose.yml`.  So, let's create a `.env` file:

```bash
# .env
POSTGRES_PASSWORD=my-secret-password
```

docker-compose automatically load `.env` and define environment variables.  Now we can modify our `docker-compose.yml` as follows:

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

Let's replace `.env` by `ssm-vault`.  We can store any secrets as well as application parameters in SSM Parameter Store:

```bash
$ ssm-vault write /app/dev/POSTGRES_USER -s
Enter text: dbuser

$ ssm-vault write /app/dev/POSTGRES_PASSWORD
Enter secret: ********
```

These values can be exposed as environment variables by `ssm-vault exec`:

```bash
$ ssm-vault exec -p /app/dev -- env | grep POSTGRES
POSTGRES_PASSWORD=my-secret-password
POSTGRES_USER=dbuser
```

Instead of creating `.env`, we can start a container like this:

```
$ ssm-vault exec -p /app/dev -- docker-compose run app
```

### Using ssm-vault within the container

It's tedious to call `ssm-vault exec` every time we start containers.  We could embed `ssm-vault` in the container and use it whenever the container starts.

You can add the following lines to your `Dockerfile` in order to install `ssm-vault`:
(See [examples/docker/Dockerfile](examples/docker/Dockerfile) for full example.)

```bash
# Install ssm-vault
ENV SSM_VAULT_VERSION v1.0.0
ENV SSM_VAULT_CHECKSUM 1f8cc1479cb5e2688eca81de4ba1ee99e4bc08a1c753b38b648a5a3bbbf4c474
ADD https://github.com/k24d/ssm-vault/releases/download/$SSM_VAULT_VERSION/ssm-vault-linux-amd64 /usr/local/bin/ssm-vault
RUN echo "$SSM_VAULT_CHECKSUM /usr/local/bin/ssm-vault" | sha256sum -c && chmod 755 /usr/local/bin/ssm-vault
```

Then, you can execute `ssm-vault exec` either by `ENTRYPOINT` or `CMD`:

```bash
# Run "/usr/local/bin/ssm-vault exec" as the entry point
ENTRYPOINT ["/usr/local/bin/ssm-vault", "exec", "-p", "/app/dev", "--"]

or

# Run "/usr/local/bin/ssm-vault exec" before the actual command
CMD ["/usr/local/bin/ssm-vault", "exec", "-p", "/app/dev", "--", "flask", "run"]
```

Before starting a container, you have to configure AWS in order to run `ssm-vault` successfully.  I'd recommend that you start a local metadata server by running [aws-vault](https://github.com/99designs/aws-vault) in server mode.  This way you can forget about AWS access keys in your container because a session token is provided by the metedata server in the same way as Amazon EC2 or ECS.

In this case, all you have to do is to start a metadata server at the beginning of day.  Then, your container will start loading secrets from SSM Parameter Store whenever you start a container:

```bash
# Start a metadata server (only once)
$ aws-vault exec development --server
...

# Start a container in a different terminal
$ docker-compose run app
```

:warning: Don't forget to set the environment variable `AWS_REGION` in your container.  A metadata server does not provide region information, so you need to specify one explicitly:

```
  app:
    environment:
      - AWS_REGION=ap-northeast-1
```
