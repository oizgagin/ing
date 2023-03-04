#!/bin/bash
set -euxo pipefail

exec redis-cli -h localhost -p 6380 --user ing_user --pass ing_pass
