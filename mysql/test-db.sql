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
