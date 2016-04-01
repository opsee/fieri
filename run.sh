#!/bin/bash
set -e

APPENV=${APPENV:-fierienv}

/opt/bin/s3kms -r us-west-1 get -b opsee-keys -o dev/$APPENV > /$APPENV

source /$APPENV && \
	/opt/bin/migrate -url "$POSTGRES_CONN" -path /migrations up && \
	/fieri
