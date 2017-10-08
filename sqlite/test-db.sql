-- sqlite example db for regression tests
-- schema must match test code's expectations

create table foo (
	id integer primary key,
	name text,
  colour nvarchar(123)
);
insert into foo (id, name, colour) values
	(1, "raaa", "blue")
;
select * from foo;
