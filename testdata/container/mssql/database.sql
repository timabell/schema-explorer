:setvar DBName "ssetest"

USE [master]

DROP DATABASE IF EXISTS [$(DBName)]

print 'Creating database $(DBName)';
CREATE DATABASE [$(DBName)];

