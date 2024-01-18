#!/bin/sh
set -e
docker exec sse-mysql mysql -pomgroot -e "drop database if exists ssetest;"
docker exec sse-mysql mysql -pomgroot -e "create database ssetest;"
docker exec -i sse-mysql mysql -pomgroot ssetest < test-db.sql

