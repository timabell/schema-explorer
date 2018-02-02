-- mssql example db for regression tests
-- schema must match test code's expectations

if object_id('DataTypeTest', 'U') is not null
begin
	drop table DataTypeTest;
end
if object_id('pet', 'U') is not null
begin
	drop table pet;
end
if object_id('toy', 'U') is not null
begin
	drop table toy;
end
if object_id('person', 'U') is not null
begin
	drop table person;
end
go
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

insert into person(personId,personName) values(1,'bob'),(2,'fred');
insert into pet(petId,petName, ownerId, favouritePersonId)values(5, 'kitty',1,2);
insert into pet(petId,petName, ownerId, favouritePersonId)values(6, 'fido',2,2);
insert into toy(toyId, toyName, belongsToId) values(11,'mouse',5);
insert into toy(toyId, toyName, belongsToId) values(12,'ball',6);
update person set favouritePetId = 5 where personId = 2;
