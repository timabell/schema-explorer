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
	belongsToId int --references pet(petId)
);

create table person (
	personId int PRIMARY KEY, personName varchar(50),
	favouritePetId int --references pet(petId)
);

create table pet (
	petId int PRIMARY KEY, petName varchar(50), "ownerId" int references person(personId),
	favouritePersonId int references person(personId)
);

alter table toy add foreign key (belongsToId) REFERENCES pet (petId);
alter table person add foreign key (favouritePetId) REFERENCES pet (petId);

insert into person(personId,personName) values(1,'bob'),(2,'fred');
insert into pet(petId,petName, "ownerId", favouritePersonId)values(5, 'kitty',1,2);
insert into pet(petId,petName, "ownerId", favouritePersonId)values(6, 'fido',2,2);
insert into toy(toyId, toyName, belongsToId) values(11,'mouse',5);
insert into toy(toyId, toyName, belongsToId) values(12,'ball',6);
update person set favouritePetId = 5 where personId = 2;

