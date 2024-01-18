#!/bin/sh
set -e
docker exec -i sse-mysql mysql -pomgroot < setup-user.sql
