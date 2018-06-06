-- mssql example db for regression tests
-- schema must match test code's expectations

/*
-- todo: drop sql doesn't cope with schema.

-- use [sdv-regression-test]; -- not supported on azure sql

-- ###################################
-- Clear out the db.
-- Verbatim from https://stackoverflow.com/a/36619064
/* Azure friendly */
/* Drop all Foreign Key constraints */
DECLARE @name VARCHAR(128)
DECLARE @constraint VARCHAR(254)
DECLARE @SQL VARCHAR(254)

SELECT @name = (SELECT TOP 1 TABLE_NAME FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS WHERE constraint_catalog=DB_NAME() AND CONSTRAINT_TYPE = 'FOREIGN KEY' ORDER BY TABLE_NAME)

WHILE @name is not null
	BEGIN
		SELECT @constraint = (SELECT TOP 1 CONSTRAINT_NAME FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS WHERE constraint_catalog=DB_NAME() AND CONSTRAINT_TYPE = 'FOREIGN KEY' AND TABLE_NAME = @name ORDER BY CONSTRAINT_NAME)
		WHILE @constraint IS NOT NULL
			BEGIN
				SELECT @SQL = 'ALTER TABLE [dbo].[' + RTRIM(@name) +'] DROP CONSTRAINT [' + RTRIM(@constraint) +']'
				EXEC (@SQL)
				PRINT 'Dropped FK Constraint: ' + @constraint + ' on ' + @name
				SELECT @constraint = (SELECT TOP 1 CONSTRAINT_NAME FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS WHERE constraint_catalog=DB_NAME() AND CONSTRAINT_TYPE = 'FOREIGN KEY' AND CONSTRAINT_NAME <> @constraint AND TABLE_NAME = @name ORDER BY CONSTRAINT_NAME)
			END
		SELECT @name = (SELECT TOP 1 TABLE_NAME FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS WHERE constraint_catalog=DB_NAME() AND CONSTRAINT_TYPE = 'FOREIGN KEY' ORDER BY TABLE_NAME)
	END
GO

/* Drop all Primary Key constraints */
DECLARE @name VARCHAR(128)
DECLARE @constraint VARCHAR(254)
DECLARE @SQL VARCHAR(254)

SELECT @name = (SELECT TOP 1 TABLE_NAME FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS WHERE constraint_catalog=DB_NAME() AND CONSTRAINT_TYPE = 'PRIMARY KEY' ORDER BY TABLE_NAME)

WHILE @name IS NOT NULL
	BEGIN
		SELECT @constraint = (SELECT TOP 1 CONSTRAINT_NAME FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS WHERE constraint_catalog=DB_NAME() AND CONSTRAINT_TYPE = 'PRIMARY KEY' AND TABLE_NAME = @name ORDER BY CONSTRAINT_NAME)
		WHILE @constraint is not null
			BEGIN
				SELECT @SQL = 'ALTER TABLE [dbo].[' + RTRIM(@name) +'] DROP CONSTRAINT [' + RTRIM(@constraint)+']'
				EXEC (@SQL)
				PRINT 'Dropped PK Constraint: ' + @constraint + ' on ' + @name
				SELECT @constraint = (SELECT TOP 1 CONSTRAINT_NAME FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS WHERE constraint_catalog=DB_NAME() AND CONSTRAINT_TYPE = 'PRIMARY KEY' AND CONSTRAINT_NAME <> @constraint AND TABLE_NAME = @name ORDER BY CONSTRAINT_NAME)
			END
		SELECT @name = (SELECT TOP 1 TABLE_NAME FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS WHERE constraint_catalog=DB_NAME() AND CONSTRAINT_TYPE = 'PRIMARY KEY' ORDER BY TABLE_NAME)
	END
GO

/* Drop all tables */
DECLARE @name VARCHAR(128)
DECLARE @SQL VARCHAR(254)

SELECT @name = (SELECT TOP 1 [name] FROM sysobjects WHERE [type] = 'U' AND category = 0 ORDER BY [name])

WHILE @name IS NOT NULL
	BEGIN
		SELECT @SQL = 'DROP TABLE [dbo].[' + RTRIM(@name) +']'
		EXEC (@SQL)
		PRINT 'Dropped Table: ' + @name
		SELECT @name = (SELECT TOP 1 [name] FROM sysobjects WHERE [type] = 'U' AND category = 0 AND [name] > @name ORDER BY [name])
	END
GO
-- ###################################
*/

drop table kitchen.sink
drop table kitchen.person
drop SCHEMA kitchen
drop table DataTypeTest
drop table toy
drop table pet
drop table person

GO
create SCHEMA kitchen;
GO
--------

create table DataTypeTest (
	intpk integer primary key,
	colCount integer,
	field_INT int,
	field_money MONEY,
	field_numeric numeric(18,7),
	field_decimal decimal(18,7),
	field_varcharmax varchar(max),
	field_nvarchar nvarchar(123),
	field_uniqueidentifier UNIQUEIDENTIFIER
);

delete DataTypeTest;
insert into DataTypeTest (
	intpk,
	colCount,
	field_INT,
	field_money,
	field_numeric,
	field_decimal,
	field_varcharmax,
	field_nvarchar,
	field_uniqueidentifier
) values (
	10, --intpk
	9, --colCount
	20, --field_INT
	1234.567, --field_money,
	987.1234500, --field_numeric,
	666.1234500, --field_decimal,
	'this is a ''text'' field',
	'blue',
	'b7a16c7a-a718-4ed8-97cb-20ccbadcc339'
),(
	11, --intpk
	0, --colCount
	-33, --field_INT
	null, --field_money,
	null, --field_numeric,
	null, --field_decimal,
	'this is a ''text'' field',
	'blue',
	'b470fa05-2111-46f9-9c97-f103b594c5f0'
)
;
--select * from DataTypeTest;

create table person (
	personId int PRIMARY KEY,
	personName nvarchar(50),
--	favouritePetId int references pet(petId)
);

create table pet (
	petId int PRIMARY KEY,
	petName nvarchar(50),
	ownerId int references person(personId),
	favouritePersonId int references person(personId)
);

create table toy (
	toyId int PRIMARY KEY,
	toyName nvarchar(50),
	belongsToId int references pet(petId)
);
alter table person add favouritePetId int references pet(petId)

-- test different schema name
/*
drop table kitchen.sink
drop table kitchen.person
drop SCHEMA kitchen
*/
create table kitchen.sink (
	sinkId int PRIMARY KEY
);
-- test a clashing name
create table kitchen.person (
	ghostPersonId int PRIMARY KEY
);

insert into person(personId,personName) values(1,'bob'),(2,'fred');
insert into pet(petId,petName, ownerId, favouritePersonId)values(5, 'kitty',1,2);
insert into pet(petId,petName, ownerId, favouritePersonId)values(6, 'fido',2,2);
insert into toy(toyId, toyName, belongsToId) values(11,'mouse',5);
insert into toy(toyId, toyName, belongsToId) values(12,'ball',6);
update person set favouritePetId = 5 where personId = 2;

create table SortFilterTest (
  id int PRIMARY KEY,
  size int,
  colour nvarchar(50),
	pattern nvarchar(50)
);
insert into SortFilterTest (id, size, colour, pattern) values
	(1, 3,'red',  'spotty'),
	(2, 4,'green','spotty'),
	(3, 2,'green','plain'),
	(4, 13,'blue', 'plain'),
	(5, 6,'blue', 'plain'),
	(6, 2,'red',  'tartan');
-- select id, size, colour, pattern from SortFilterTest ;
-- select '---';
-- select id, size, colour, pattern from SortFilterTest where pattern = 'plain' order by colour, size desc;
