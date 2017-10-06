-- mssql example db for regression tests
-- schema must match test code's expectations

if object_id('foo', 'U') is not null
begin
	drop table foo;
end

create table foo (
	id integer primary key,
	name varchar(max)
);
insert into foo (id, name) values
	(1, 'raaa'),
	(2, 'meow')
;
select * from foo;
