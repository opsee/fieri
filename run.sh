#!/bin/bash
set -e

# relying on set -e to catch errors?
/opt/bin/ec2-env > /ec2env
eval "$(< /ec2env)"
/opt/bin/s3kms get -b opsee-keys -o dev/fierienv > /fierienv

source /fierienv && \
	/opt/bin/migrate -url "$POSTGRES_CONN" -path /migrations up && \
	/fieri
