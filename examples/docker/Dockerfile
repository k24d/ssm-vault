FROM python:3.7

WORKDIR /app

COPY ./requirements.txt /app/requirements.txt

RUN pip install -r requirements.txt

# Install ssm-vault
ENV SSM_VAULT_VERSION v1.0.0
ENV SSM_VAULT_CHECKSUM 1f8cc1479cb5e2688eca81de4ba1ee99e4bc08a1c753b38b648a5a3bbbf4c474
RUN wget -O /usr/local/bin/ssm-vault https://github.com/k24d/ssm-vault/releases/download/$SSM_VAULT_VERSION/ssm-vault-linux-amd64
RUN echo "$SSM_VAULT_CHECKSUM /usr/local/bin/ssm-vault" | sha256sum -c && chmod 755 /usr/local/bin/ssm-vault

ENTRYPOINT ["/usr/local/bin/ssm-vault", "exec", "-p", "/app/dev", "--"]

ENV FLASK_APP app.py
CMD ["flask", "run"]
