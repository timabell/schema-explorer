-- sqlite example db for regression tests
-- schema must match test code's expectations

drop table if exists DataTypeTest;

create table DataTypeTest (
	intpk integer primary key,
	colCount int,
	field_INT INT
);

insert into DataTypeTest(
	intpk,
	colCount,
	-- https://www.sqlite.org/datatype3.html#affinity_name_examples
	field_INT
)values(
	10, --intpk
	3, --colCount
	20 --INT
),(
	11, --intpk
	0, --colCount
	-33 --INT
);

-- select * from DataTypeTest;

/*
create table toy (
	toyId int PRIMARY KEY,
	toyName nvarchar(50),
	belongsToId int references pet(petId)
);

create table person (
	personId int PRIMARY KEY,
	personName nvarchar(50),
	favouritePetId int references pet(petId)
);

create table pet (
	petId int PRIMARY KEY,
	petName nvarchar(50),
	ownerId int references person(personId),
	favouritePersonId int references person(personId)
);

insert into person(personId,personName) values(1,'bob'),(2,'fred');
insert into pet(petId,petName, ownerId, favouritePersonId)values(5, 'kitty',1,2);
insert into pet(petId,petName, ownerId, favouritePersonId)values(6, 'fido',2,2);
insert into toy(toyId, toyName, belongsToId) values(11,'mouse',5);
insert into toy(toyId, toyName, belongsToId) values(12,'ball',6);
update person set favouritePetId = 5 where personId = 2;
*/

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
