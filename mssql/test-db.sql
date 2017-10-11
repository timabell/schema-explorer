-- mssql example db for regression tests
-- schema must match test code's expectations

if object_id('DataTypeTest', 'U') is not null
begin
	drop table DataTypeTest;
end

create table DataTypeTest (
	intpk integer primary key,
	varcharmax varchar(max),
	nvarchar nvarchar(123)
);
insert into DataTypeTest (id, name, colour) values
	(1, 'this is a ''text'' field', 'blue')
;
select * from DataTypeTest;
