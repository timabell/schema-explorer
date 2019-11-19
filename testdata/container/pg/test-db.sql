-- postgres example db for regression tests
-- schema must match test code's expectations


-- todo: commented out types
create table "DataTypeTest" (
  intpk integer primary key,
	"col_count" int,
	-- numeric
-- 	"field_bare_int" INT, -- reported as int4, so not named field_int to avoiding clash with sqlite int
-- 	"field_int4" INT4,
	"field_null_int" int null,
-- 	"field_pg_smallint" smallint,
-- 	"field_interval" interval null,
-- 	"field_float8" double precision null,
-- 	"field_money" money null,
-- 	"field_numeric" numeric(5,2) null,
-- 	"field_float4" real null,
	-- char
-- 	"field_bpchar" character(20) null,
-- 	"field_varchar" varchar(20) null,
	"field_text" text null,
	-- bool
-- 	"field_bool" boolean null,
-- 	"field_bit" bit null,
-- 	"field_varbit" varbit null,
	--binary
-- 	"field_bytea" bytea null,
	-- date
-- 	"field_date" date null,
-- 	"field_time" time null,
-- 	"field_timetz" timetz null,
-- 	"field_timestamp" timestamp null,
-- 	"field_timestamptz" timestamptz null,
	-- geom
-- 	"field_circle" circle null,
-- 	"field_line" line null,
-- 	"field_lseg" lseg null,
-- 	"field_path" path null,
-- 	"field_point" point null,
-- 	"field_polygon" polygon null,
	-- networking
-- 	"field_inet" inet null, -- ip4/6 network address
-- 	"field_cidr" cidr null, -- ip4/6 host address
-- 	"field_macaddr" macaddr null,
	-- "field_macaddr8" macaddr8 null, -- ERROR:  type "macaddr8" does not exist
	-- json
	"field_json" json,
	"field_jsonb" jsonb,
	-- misc
-- 	"field_pg_lsn" pg_lsn,
-- 	"field_smallserial" smallserial,
-- 	"field_serial" serial,
-- 	"field_tsquery" tsquery, -- text-search query
-- 	"field_tsvector" tsvector, -- text-search document
-- 	"field_txid_snapshot" txid_snapshot, -- user-level transaction ID snapshot
-- 	"field_uuid" uuid,
-- 	"field_xml" xml,
	-- null
	"field_not_null_int" int not null
);

insert into "DataTypeTest"(
  intpk,
  col_count,
--   field_bare_int,
--   field_int4,
  field_null_int,
--   field_pg_smallint,
--   field_interval,
--   field_float8,
--   field_money,
--   field_numeric,
--   field_float4,
--   field_bpchar,
--   field_varchar,
  field_text,
--   field_bool,
--   field_bit,
--   field_varbit,
--   field_bytea,
--   field_date,
--   field_time,
--   field_timetz,
--   field_timestamp,
--   field_timestamptz,
--   field_circle,
--   field_line,
--   field_lseg,
--   field_path,
--   field_point,
--   field_polygon,
--   field_inet,
--   field_cidr,
--   field_macaddr,
  field_json,
  field_jsonb,
--   field_pg_lsn,
--   field_smallserial,
--   field_serial,
--   field_tsquery,
--   field_tsvector,
--   field_txid_snapshot,
--   field_uuid,
--   field_xml,
  field_not_null_int
) values (
10, -- intpk,
7, -- col_count,
-- 20, -- field_bare_int,
-- 1984, -- field_int4,
null, -- field_null_int,
-- null, -- field_pg_smallint,
-- null, -- field_interval,
-- null, -- field_float8,
-- null, -- field_money,
-- 987.12345, -- field_numeric,
-- null, -- field_float4,
-- null, -- field_bpchar,
-- null, -- field_varchar,
'a_TEXT', -- field_text,
-- null, -- field_bool,
-- null, -- field_bit,
-- null, -- field_varbit,
-- null, -- field_bytea,
-- '1984-04-02', -- field_date,
-- null, -- field_time,
-- null, -- field_timetz,
-- null, -- field_timestamp,
-- null, -- field_timestamptz,
-- null, -- field_circle,
-- null, -- field_line,
-- null, -- field_lseg,
-- null, -- field_path,
-- null, -- field_point,
-- null, -- field_polygon,
-- null, -- field_inet,
-- null, -- field_cidr,
-- null, -- field_macaddr,
'[{"name": "frank"}, {"name": "sinatra"}]'::json, -- field_json,
'[{"name": "frank"}, {"name": "sinatra"}]'::jsonb, -- field_jsonb,
-- null, -- field_pg_lsn,
-- 1234, -- field_smallserial, // seems to not be nullable
-- 1243456, -- field_serial,
-- null, -- field_tsquery,
-- null, -- field_tsvector,
-- null, -- field_txid_snapshot,
-- null, -- field_uuid,
-- null, -- field_xml,
4541 -- field_not_null_int)
),(
11, -- intpk,
0, -- col_count,
-- -33, -- field_bare_int,
-- 1978, -- field_int4,
null, -- field_null_int,
-- null, -- field_pg_smallint,
-- null, -- field_interval,
-- null, -- field_float8,
-- null, -- field_money,
-- null, -- field_numeric,
-- null, -- field_float4,
-- null, -- field_bpchar,
-- null, -- field_varchar,
null, -- field_text,
-- null, -- field_bool,
-- null, -- field_bit,
-- null, -- field_varbit,
-- null, -- field_bytea,
-- null, -- field_date,
-- null, -- field_time,
-- null, -- field_timetz,
-- null, -- field_timestamp,
-- null, -- field_timestamptz,
-- null, -- field_circle,
-- null, -- field_line,
-- null, -- field_lseg,
-- null, -- field_path,
-- null, -- field_point,
-- null, -- field_polygon,
-- null, -- field_inet,
-- null, -- field_cidr,
-- null, -- field_macaddr,
null, -- field_json,
null, -- field_jsonb,
-- null, -- field_pg_lsn,
-- 12345, -- field_smallserial,
-- 12434567, -- field_serial,
-- null, -- field_tsquery,
-- null, -- field_tsvector,
-- null, -- field_txid_snapshot,
-- null, -- field_uuid,
-- null, -- field_xml,
4542 -- field_not_null_int)
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

create table "FkParent"(
  "parentPk" int primary key
);
create table "FkChild"(
  id int primary key,
  "parentId" int references "FkParent"("parentPk")
);

insert into "FkParent"("parentPk") values(10);
insert into "FkParent"("parentPk") values(11);
insert into "FkParent"("parentPk") values(12);
insert into "FkChild"(id, "parentId") values(100,10);
insert into "FkChild"(id, "parentId") values(101,10);
insert into "FkChild"(id, "parentId") values(102,10);
insert into "FkChild"(id, "parentId") values(110,11);
insert into "FkChild"(id, "parentId") values(111,11);
insert into "FkChild"(id, "parentId") values(112,11);

create table index_test(
  id int primary key,
  has_index varchar(10),
  compound_a varchar(10),
  compound_b varchar(10),
  complex_index varchar(10),
  unique_index varchar(10)
);
create index "IX_on_has_index" on index_test (has_index);
create index "IX_compound" on index_test (compound_a, compound_b);
create index "IX_complex" on index_test (lower(complex_index)); -- this won't show in the column's index list but will show in the table/database list
create unique index "IX_unique" on index_test (unique_index);

create table analysis_test(
  colour varchar(50)
);
insert into analysis_test(colour)values
('red'), ('red'), ('red'),
('blue'), ('blue'),
('green'),
(null), (null), (null), (null);

-- check keywords are escaped by making a nasty schema/table/column name
create schema "identity";
create table "identity"."select" (
  id int primary key,
  "table" varchar(50)
);
insert into "identity"."select" (id, "table") values (1, 'times');

-- select * from "identity"."select";

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
