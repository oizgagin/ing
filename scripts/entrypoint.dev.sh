#!/bin/bash

set -eux

/usr/local/bin/wait-for-it.sh broker:9092
/usr/local/bin/wait-for-it.sh postgres:5432
/usr/local/bin/wait-for-it.sh redis1:6379
/usr/local/bin/wait-for-it.sh redis2:6379
/usr/local/bin/wait-for-it.sh redis3:6379
exec /usr/local/bin/ing
