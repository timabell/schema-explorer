-- postgres example db for regression tests
-- schema must match test code's expectations

create table "DataTypeTest" (
  intpk integer primary key,
	"colCount" int,
	"field_INT4" INT,
	"field_NotNullInt" int not null,
	"field_NullInt" int null
);

insert into "DataTypeTest"( intpk, "colCount", "field_INT4", "field_NotNullInt"
)values(
	10, --intpk
	5, --colCount
	20, --INT
	1984
),(
	11, --intpk
	0, --colCount
	-33, --INT
	1978
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
	padding int,
	"colA" varchar(10),
	"colB" varchar(10),
	"badger" varchar(50),
	primary key ("colA", "colB")
);

create table "CompoundKeyAunty"(
	id int,
	"colB" varchar(10),
	primary key ("colB")
);

create table "CompoundKeyChild"(
	id int PRIMARY KEY,
	"colA" varchar(10),
	"colB" varchar(10),
  noise varchar(50),
	foreign key ("colB") references "CompoundKeyAunty"("colB"),
	foreign key ("colA", "colB") references "CompoundKeyParent"("colA", "colB")
);

insert into "CompoundKeyParent"("id", "colA", "colB", "badger") values
	(1,'a1', 'b1', 'mash'),
	(2,'a2', 'b2', 'bodger'),
	(3,'a2', 'b3', 'mmmmm'),
	(4,'a<&''2\6', 'b2', 'mwah ha ha');
insert into "CompoundKeyAunty"(id, "colB")values
	(10, 'b1'),
	(11, 'b2'),
	(12, 'b3');
insert into "CompoundKeyChild"("id", "colA", "colB", "noise") values
	(1,'a1', 'b1', 'pig'),
	(2,'a1', 'b1', 'swine'),
	(3,'a2', 'b2', 'horse'),
	(4,'a<&''2\6', 'b2', 'does it blend?');
