-- sqlite example db for regression tests
-- schema must match test code's expectations

drop table if exists DataTypeTest;

create table DataTypeTest (
	intpk integer primary key,
	col_count int,
	field_int int,
	field_integer integer,
	field_tinyint tinyint,
	field_smallint smallint,
	field_mediumint mediumint,
	field_bigint bigint,
	field_int2 int2,
	field_int8 int8,
	field_character character(20),
	field_sqlite_varchar varchar(255),
	field_nchar nchar(55),
	field_nvarchar nvarchar(100),
	field_text text,
	field_blob blob,
	field_real real,
	field_double double,
	field_doubleprecision double precision,
	field_float float,
	field_numeric numeric,
	field_sqlite_decimal decimal(10,5),
	field_boolean boolean,
	field_date date,
	field_datetime datetime,
	field_not_null_int int not null,
	field_null_int int null
);

insert into DataTypeTest(
	intpk,
	col_count,
	field_int,
	field_integer,
	field_tinyint,
	field_smallint,
	field_mediumint,
	field_bigint,
	field_int2,
	field_int8,
	field_character,
	field_sqlite_varchar,
	field_nchar,
	field_nvarchar,
	field_text,
	field_blob,
	field_real,
	field_double,
	field_doubleprecision,
	field_float,
	field_numeric,
	field_sqlite_decimal,
	field_boolean,
	field_date,
	field_datetime,
	field_not_null_int
)values(
	10, -- intpk
	27, -- col_count
	20, -- INT
	30, -- INTEGER
	50, -- TINYINT
	60, -- SMALLINT
	70, -- MEDIUMINT
	80, -- BIGINT
	100, -- INT2
	110, -- INT8
	'a_CHARACTER', -- CHARACTER
	'a_VARCHAR', -- sqlite_VARCHAR
	'a_NCHAR', -- NCHAR
	'a_NVARCHAR', -- NVARCHAR
	'a_TEXT', -- TEXT
	'a_BLOB', -- BLOB
	1.234, -- REAL
	1.234, -- DOUBLE
	1.234, -- DOUBLE PRECISION
	1.234, -- FLOAT
	987.12345, -- NUMERIC
	1.234, -- DECIMAL
	1, -- BOOLEAN
	'1984-04-02', -- DATE
	'1984-04-02 11:12', -- DATETIME'
	1984
),(
	11, -- intpk
	0, -- col_count
	-33, -- INT
	30, -- INTEGER
	50, -- TINYINT
	60, -- SMALLINT
	70, -- MEDIUMINT
	80, -- BIGINT
	100, -- INT2
	110, -- INT8
	null, -- CHARACTER
	null, -- sqlite_VARCHAR
	null, -- NCHAR
	null, -- NVARCHAR
	null, -- TEXT
	null, -- BLOB
	null, -- REAL
	null, -- DOUBLE
	null, -- DOUBLE PRECISION
	null, -- FLOAT
	null, -- NUMERIC
	null, -- DECIMAL
	0, -- BOOLEAN
	null, -- DATE
	null, -- DATETIME'
	1978
);
