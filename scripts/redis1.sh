#!/bin/bash
set -euxo pipefail

exec redis-cli -h localhost -p 6379 --user ing_user --pass ing_pass
