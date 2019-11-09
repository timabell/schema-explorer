#!/bin/sh
docker exec -i sse-mysql mysql -pomgroot < setup-user.sql
