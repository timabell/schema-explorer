-- mssql example db for regression tests
-- schema must match test code's expectations

if object_id('DataTypeTest', 'U') is not null
begin
	drop table DataTypeTest;
end
-- split here
create table DataTypeTest (
	intpk integer primary key,
	colCount integer,
	field_INT int,
	field_varcharmax varchar(max),
	field_nvarchar nvarchar(123)
);

delete DataTypeTest;
insert into DataTypeTest (
	intpk,
	colCount,
	field_INT,
	field_varcharmax,
	field_nvarchar
) values (
	10, --intpk
	5, --colCount
	20, --field_INT
	'this is a ''text'' field',
	'blue'
),(
	11, --intpk
	0, --colCount
	-33, --field_INT
	'this is a ''text'' field',
	'blue'
)
;
select * from DataTypeTest;
