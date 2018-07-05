-- postgres example db for regression tests
-- schema must match test code's expectations

create table "DataTypeTest" ( intpk integer primary key, "colCount" int, "field_INT4" INT );

insert into "DataTypeTest"( intpk, "colCount", "field_INT4"
)values(
	10, --intpk
	3, --colCount
	20 --INT
),(
	11, --intpk
	0, --colCount
	-33 --INT
);

create table toy (
	toyId int PRIMARY KEY, toyName varchar(50),
	belongsToId int --references pet("petId")
);

create table person (
	personId int PRIMARY KEY, personName varchar(50),
	favouritePetId int --references pet("petId")
);

create table pet (
	"petId" int PRIMARY KEY, petName varchar(50), "ownerId" int references person(personId),
	favouritePersonId int references person(personId)
);

alter table toy add foreign key (belongsToId) REFERENCES pet ("petId");
alter table person add foreign key (favouritePetId) REFERENCES pet ("petId");

insert into person(personId,personName) values(1,'bob'),(2,'fred');
insert into pet("petId",petName, "ownerId", favouritePersonId)values(5, 'kitty',1,2);
insert into pet("petId",petName, "ownerId", favouritePersonId)values(6, 'fido',2,2);
insert into toy(toyId, toyName, belongsToId) values(11,'mouse',5);
insert into toy(toyId, toyName, belongsToId) values(12,'ball',6);
update person set favouritePetId = 5 where personId = 2;


-- sort-filter testing

create table "SortFilterTest" (
  id int PRIMARY KEY,
  size int,
  colour varchar(50),
	pattern varchar(50)
);

insert into "SortFilterTest" (id, size, colour, pattern) values
	(1, 3,  'red',   'spotty'),
	(2, 4,  'green', 'spotty'),
	(3, 2,  'green', 'plain'),
	(4, 21, 'blue',  'plain'),
	(5, 23, 'blue',  'plain'),
	(6, 22, 'blue',  'plain'),
	(7, 2,  'red',   'tartan');

-- select * from "SortFilterTest" ;
-- select '---';
-- -- this is what the test should run:
-- select * from "SortFilterTest" where pattern = 'plain' order by colour, size desc;

create table "CompoundKeyParent"(
	id int,
	"colA" varchar(10),
	"colB" varchar(10),
	primary key ("colA", "colB")
);

create table "CompoundKeyChild"(
	id int PRIMARY KEY,
	"colA" varchar(10),
	"colB" varchar(10),
  noise varchar(50),
	foreign key ("colA", "colB") references "CompoundKeyParent"("colA", "colB")
);
