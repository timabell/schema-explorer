-- sqlite example db for regression tests
-- schema must match test code's expectations

create table foo (
	id integer primary key,
	name text
);
insert into foo (id, name) values
	(1, "raaa"),
	(2, "meow")
;
select * from foo;
