
-- mysql example db for regression tests
-- schema must match test code's expectations

drop table if exists DataTypeTest;

create table DataTypeTest (
	intpk integer primary key,
	col_count int,
	field_mysql_int int,
	field_mysql_character character(20),
	field_mysql_nchar nchar(55),
	field_mysql_nvarchar nvarchar(100),
	field_text text,
	field_blob blob,
	field_mysql_real real,
	field_double double,
	field_mysql_doubleprecision double precision,
	field_float float,
	field_mysql_boolean boolean,
	field_not_null_int int not null,
	field_null_int int null
);

insert into DataTypeTest(
	intpk,
	col_count,
	field_mysql_int,
	field_mysql_character,
	field_mysql_nchar,
	field_mysql_nvarchar,
	field_text,
	field_blob,
	field_mysql_real,
	field_double,
	field_mysql_doubleprecision,
	field_float,
	field_mysql_boolean,
	field_not_null_int
)values(
	10, -- intpk
	15, -- col_count
	20, -- INT
	'a_CHARACTER', -- CHARACTER
	'a_NCHAR', -- NCHAR
	'a_NVARCHAR', -- NVARCHAR
	'a_TEXT', -- TEXT
	'a_BLOB', -- BLOB
	1.234, -- REAL
	1.234, -- DOUBLE
	1.234, -- DOUBLE PRECISION
	1.234, -- FLOAT
	1, -- BOOLEAN
	1984
),(
	11, -- intpk
	0, -- col_count
	-33, -- INT
	null, -- CHARACTER
	null, -- NCHAR
	null, -- NVARCHAR
	null, -- TEXT
	null, -- BLOB
	null, -- REAL
	null, -- DOUBLE
	null, -- DOUBLE PRECISION
	null, -- FLOAT
	0, -- BOOLEAN
	1978
);

-- select * from DataTypeTest;

create table person (
	personId int PRIMARY KEY,
	personName nvarchar(50),
	favouritePetId int
);

create table pet (
	petId int PRIMARY KEY,
	petName nvarchar(50),
	ownerId int,
	favouritePersonId int,
	foreign key (ownerId) references person(personId),
	foreign key (favouritePersonId) references person(personId)
);

create table toy (
	toyId int PRIMARY KEY,
	toyName nvarchar(50),
	belongsToId int,
	foreign key (belongsToId) references pet(petId)
);

alter table person
 add foreign key (favouritePetId) references pet(petId);

insert into person(personId,personName) values(1,'bob'),(2,'fred');
insert into pet(petId,petName, ownerId, favouritePersonId)values(5, 'kitty',1,2);
insert into pet(petId,petName, ownerId, favouritePersonId)values(6, 'fido',2,2);
insert into toy(toyId, toyName, belongsToId) values(11,'mouse',5);
insert into toy(toyId, toyName, belongsToId) values(12,'ball',6);
update person set favouritePetId = 5 where personId = 2;

/* for manual testing of diagrams, commented out to avoid interfering with regression test

-- tall, for checking diagram scaling manually
create table up1 ( upId int PRIMARY KEY);
create table up2 ( upId int PRIMARY KEY, anotherId int references up1(upId));
create table up3 ( upId int PRIMARY KEY, anotherId int references up2(upId));
create table up4 ( upId int PRIMARY KEY, anotherId int references up3(upId));
create table up5 ( upId int PRIMARY KEY, anotherId int references up4(upId));
create table up6 ( upId int PRIMARY KEY, anotherId int references up5(upId));
create table up7 ( upId int PRIMARY KEY, anotherId int references up6(upId));
create table up8 ( upId int PRIMARY KEY, anotherId int references up7(upId));
create table up9 ( upId int PRIMARY KEY, anotherId int references up8(upId));
create table up10 ( upId int PRIMARY KEY, anotherId int references up9(upId));
create table up11 ( upId int PRIMARY KEY, anotherId int references up10(upId));
create table up12 ( upId int PRIMARY KEY, anotherId int references up11(upId));
create table up13 ( upId int PRIMARY KEY, anotherId int references up12(upId));
create table up14 ( upId int PRIMARY KEY, anotherId int references up13(upId));
create table up15 ( upId int PRIMARY KEY, anotherId int references up14(upId));
create table up16 ( upId int PRIMARY KEY, anotherId int references up15(upId));
create table up17 ( upId int PRIMARY KEY, anotherId int references up16(upId));
create table up18 ( upId int PRIMARY KEY, anotherId int references up17(upId));
create table up19 ( upId int PRIMARY KEY, anotherId int references up18(upId));
create table up20 ( upId int PRIMARY KEY, anotherId int references up19(upId));

-- wide, for checking diagram scaling manually
create table parent ( parentId int PRIMARY KEY, grandParentId int references up20(upId));
create table child01 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child02 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child03 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child04 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child05 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child06 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child07 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child08 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child09 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child10 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child11 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child12 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child13 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child14 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child15 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child16 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child17 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child18 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child19 ( childId int PRIMARY KEY, parentId int references parent(parentId));
create table child20 ( childId int PRIMARY KEY, parentId int references parent(parentId));
*/

-- sort-filter testing

create table SortFilterTest (
  id int PRIMARY KEY,
  size int,
  colour nvarchar(50),
	pattern nvarchar(50)
);

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
  parentId int,
  foreign key (parentId) references FkParent(parentPk)
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

create table index_test(
  id int primary key,
  has_index varchar(10),
  compound_a varchar(10),
  compound_b varchar(10),
  complex_index varchar(10),
  unique_index varchar(10)
);
create index IX_on_has_index on index_test (has_index);
create index IX_compound on index_test (compound_a, compound_b);
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
create table `select` (
  id int primary key,
  `table` varchar(50)
);
insert into `select` (id, `table`) values (1, 'times');

-- select * from `select`;

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
  dumb_filter varchar(10) default 'filtration',
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
