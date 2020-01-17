#!/bin/bash

PARAMETER_PATH="/app/${APP_ENV:-dev}"

/usr/local/bin/ssm-vault render /app/config.template -p $PARAMETER_PATH -o /app/config.py

export FLASK_APP=/app/app.py
exec flask run
