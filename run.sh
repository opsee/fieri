#!/bin/bash
set -e

APPENV=${APPENV:-fierienv}

# relying on set -e to catch errors?
/opt/bin/ec2-env > /ec2env
eval "$(< /ec2env)"
/opt/bin/s3kms get -b opsee-keys -o dev/$APPENV > /$APPENV

source /$APPENV && \
	/opt/bin/migrate -url "$POSTGRES_CONN" -path /migrations up && \
	/fieri
