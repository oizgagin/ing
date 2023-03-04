#!/bin/bash
set -euxo pipefail

exec env PGPASSWORD=ing_pass psql -h localhost -p 5432 -U ing_user ing
