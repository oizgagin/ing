#!/bin/bash
set -euxo pipefail

exec redis-cli -h localhost -p 6381 --user ing_user --pass ing_pass
