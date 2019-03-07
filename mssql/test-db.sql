-- mssql example db for regression tests
-- schema must match test code's expectations

-- print 'creating schema...'
go
create SCHEMA kitchen;

go
-- user had a schema named the same as a keyword. good test for escaping correctly
create SCHEMA [identity];
go

-- print 'creating tables etc...'
set nocount on;
--------

create table DataTypeTest (
	intpk integer primary key,
	col_count integer,
	field_int int,
-- 	field_money MONEY,
-- 	field_numeric numeric(18,7),
-- 	field_decimal decimal(18,7),
-- 	field_varcharmax varchar(max),
-- 	field_nvarchar nvarchar(123),
	field_uniqueidentifier UNIQUEIDENTIFIER,
	field_not_null_int int not null,
	field_null_int int null
);

delete DataTypeTest;
insert into DataTypeTest (
	intpk,
	col_count,
	field_int,
-- 	field_money,
-- 	field_numeric,
-- 	field_decimal,
-- 	field_varcharmax,
-- 	field_nvarchar,
	field_uniqueidentifier,
	field_not_null_int
) values (
	10, --intpk
	6, --col_count
	20, --field_int
-- 	1234.567, --field_money,
-- 	987.1234500, --field_numeric,
-- 	666.1234500, --field_decimal,
-- 	'this is a ''text'' field', -- nvarcharmax
-- 	'a_NVARCHAR',
	'b7a16c7a-a718-4ed8-97cb-20ccbadcc339',
	1984
),(
	11, --intpk
	0, --col_count
	-33, --field_int
-- 	null, --field_money,
-- 	null, --field_numeric,
-- 	null, --field_decimal,
-- 	'this is a ''text'' field', -- nvarcharmax
-- 	'blue', -- nvarchar
	'b470fa05-2111-46f9-9c97-f103b594c5f0',
	1978
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
alter table person add favouritePetId int references pet(petId);

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


-- sort-filter testing

create table SortFilterTest (
  id int PRIMARY KEY,
  size int,
  colour nvarchar(50),
	pattern nvarchar(50)
);

delete from SortFilterTest;
insert into SortFilterTest (id, size, colour, pattern) values
	(1, 3,  'red',   'spotty'),
	(2, 4,  'green', 'spotty'),
	(3, 2,  'green', 'plain'),
	(4, 21, 'blue',  'plain'),
	(5, 23, 'blue',  'plain'),
	(6, 22, 'blue',  'plain'),
	(7, 2,  'red',   'tartan');

-- select id, size, colour, pattern from SortFilterTest ;
-- select '---';
-- select id, size, colour, pattern from SortFilterTest where pattern = 'plain' order by colour, size desc;

create table CompoundKeyParent(
	id int,
	padding int,
	colA varchar(10),
	colB varchar(10),
	badger varchar(50),
	primary key (colA, colB)
);

create table CompoundKeyAunty(
	id int,
	colB varchar(10),
	primary key (colB)
);

create table CompoundKeyChild(
	id int PRIMARY KEY,
	colA varchar(10),
	colB varchar(10),
	noise varchar(50),
	foreign key (colB) references CompoundKeyAunty(colB),
	foreign key (colA, colB) references CompoundKeyParent(colA, colB)
);

insert into CompoundKeyParent(id, colA, colB, badger)values
	(1,'a1', 'b1', 'mash'),
	(2,'a2', 'b2', 'bodger'),
	(3,'a2', 'b3', 'mmmmm'),
	(4,'a<&''2\6', 'b2', 'mwah ha ha');
insert into CompoundKeyAunty(id, colB)values
	(10, 'b1'),
	(11, 'b2'),
	(12, 'b3');
insert into CompoundKeyChild(id, colA, colB, noise)values
	(1,'a1', 'b1', 'pig'),
	(2,'a1', 'b1', 'swine'),
	(3,'a2', 'b2', 'horse'),
	(4,'a<&''2\6', 'b2', 'does it blend?');

create table FkParent(
  parentPk int primary key
);
create table FkChild(
  id int primary key,
  parentId int references FkParent(parentPk)
);

insert into FkParent(parentPk) values(10);
insert into FkParent(parentPk) values(11);
insert into FkParent(parentPk) values(12);
insert into FkChild(id, parentId) values(100,10);
insert into FkChild(id, parentId) values(101,10);
insert into FkChild(id, parentId) values(102,10);
insert into FkChild(id, parentId) values(110,11);
insert into FkChild(id, parentId) values(111,11);
insert into FkChild(id, parentId) values(112,11);

-- drop table index_test;
create table index_test(
  id int primary key,
  has_index varchar(10),
  compound_a varchar(10),
  compound_b varchar(10),
  complex_index varchar(10),
  unique_index varchar(10)
--   lower_complex as lower(complex_index) persisted // todo: didn't work on docker mssql
);

create index IX_on_has_index on index_test (has_index);
create index IX_compound on index_test (compound_a, compound_b);
-- create index IX_complex on index_test (lower_complex);
create unique index IX_unique on index_test (unique_index);

create table analysis_test(
  colour varchar(50)
);
insert into analysis_test(colour)values
('red'), ('red'), ('red'),
('blue'), ('blue'),
('green'),
(null), (null), (null), (null);

-- check keywords are escaped by making a nasty schema/table/column name
create table [identity].[select] (
  id int primary key identity,
  [table] varchar(50)
);
insert into [identity].[select] ([table]) values ('times');

-- select * from [identity].[select];

create table poke(
  id int primary key,
  name varchar(10),
  dumb_filter varchar(10) -- name clash to induce error if not qualified with table name/alias
);
insert into poke (id, name) values (11, 'piggy');
insert into poke (id, name) values (12, null);
insert into poke (id, name) values (13, 'pie');


create table peek(
  id int primary key,
  something varchar(10),
  dumb_filter varchar(10) default ('filtration'),
  poke_id int,
  pike_id int,
  foreign key (poke_id) references poke(id)
);

insert into peek (id, something, poke_id) values (1, 'wiggy', 11);
insert into peek (id, something, poke_id) values (2, 'weggy', 12);
insert into peek (id, something, poke_id) values (3, 'woggy', null);
insert into peek (id, something, poke_id) values (4, 'wibble', 12);

create table coz(
  id int primary key,
  name varchar(10),
  poke_id int,
  foreign key (poke_id) references poke(id)
);
insert into coz(id, name, poke_id) values (1, 'andy', 11);
insert into coz(id, name, poke_id) values (2, 'bob', 11);
