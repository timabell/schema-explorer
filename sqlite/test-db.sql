-- sqlite example db for regression tests
-- schema must match test code's expectations

drop table if exists DataTypeTest;

create table DataTypeTest (
	intpk integer primary key,
	colCount int,
	field_INT INT,
	field_INTEGER INTEGER,
	field_TINYINT TINYINT,
	field_SMALLINT SMALLINT,
	field_MEDIUMINT MEDIUMINT,
	field_BIGINT BIGINT,
	field_UNSIGNED UNSIGNED BIG INT,
	field_INT2 INT2,
	field_INT8 INT8,
	field_CHARACTER CHARACTER(20),
	field_VARCHAR VARCHAR(255),
	field_VARYING VARYING CHARACTER(255),
	field_NCHAR NCHAR(55),
	field_NATIVE NATIVE CHARACTER(70),
	field_NVARCHAR NVARCHAR(100),
	field_TEXT TEXT,
	field_CLOB CLOB,
	field_BLOB BLOB,
	field_REAL REAL,
	field_DOUBLE DOUBLE,
	field_DOUBLEPRECISION DOUBLE PRECISION,
	field_FLOAT FLOAT,
	field_NUMERIC NUMERIC,
	field_DECIMAL DECIMAL(10,5),
	field_BOOLEAN BOOLEAN,
	field_DATE DATE,
	field_DATETIME DATETIME
);

insert into DataTypeTest(
	intpk,
	colCount,
	-- https://www.sqlite.org/datatype3.html#affinity_name_examples
	field_INT,
	field_INTEGER,
	field_TINYINT,
	field_SMALLINT,
	field_MEDIUMINT,
	field_BIGINT,
	field_UNSIGNED,
	field_INT2,
	field_INT8,
	field_CHARACTER,
	field_VARCHAR,
	field_VARYING,
	field_NCHAR,
	field_NATIVE,
	field_NVARCHAR,
	field_TEXT,
	field_CLOB,
	field_BLOB,
	field_REAL,
	field_DOUBLE,
	field_DOUBLEPRECISION,
	field_FLOAT,
	field_NUMERIC,
	field_DECIMAL,
	field_BOOLEAN,
	field_DATE,
	field_DATETIME
)values(
	10, --intpk
	29, --colCount
	20, --INT
	30, --INTEGER
	50, --TINYINT
	60, --SMALLINT
	70, --MEDIUMINT
	80, --BIGINT
	90, --UNSIGNED
	100, --INT2
	110, --INT8
	'field_CHARACTER', --CHARACTER
	'field_VARCHAR', --VARCHAR
	'field_VARYING', --VARYING
	'field_NCHAR', --NCHAR
	'field_NATIVE', --NATIVE
	'field_NVARCHAR', --NVARCHAR
	'field_TEXT', --TEXT
	'field_CLOB', --CLOB
	'field_BLOB', --BLOB
	'field_REAL', --REAL
	'field_DOUBLE', --DOUBLE
	'field_DOUBLEPRECISION', --DOUBLE PRECISION
	'field_FLOAT', --FLOAT
	'field_NUMERIC', --NUMERIC
	'field_DECIMAL', --DECIMAL
	1, --BOOLEAN
	'field_DATE', --DATE
	'field_DATETIME' --DATETIME'
),(
	11, --intpk
	0, --colCount
	-33, --INT
	30, --INTEGER
	50, --TINYINT
	60, --SMALLINT
	70, --MEDIUMINT
	80, --BIGINT
	90, --UNSIGNED
	100, --INT2
	110, --INT8
	'field_CHARACTER', --CHARACTER
	'field_VARCHAR', --VARCHAR
	'field_VARYING', --VARYING
	'field_NCHAR', --NCHAR
	'field_NATIVE', --NATIVE
	'field_NVARCHAR', --NVARCHAR
	'field_TEXT', --TEXT
	'field_CLOB', --CLOB
	'field_BLOB', --BLOB
	'field_REAL', --REAL
	'field_DOUBLE', --DOUBLE
	'field_DOUBLEPRECISION', --DOUBLE PRECISION
	'field_FLOAT', --FLOAT
	'field_NUMERIC', --NUMERIC
	'field_DECIMAL', --DECIMAL
	1, --BOOLEAN
	'field_DATE', --DATE
	'field_DATETIME' --DATETIME'
);

-- select * from DataTypeTest;

create table person (
	personId int PRIMARY KEY,
	personName nvarchar(50)
);

create table pet (
	petId int PRIMARY KEY,
	ownerId int references person(personId),
	petName nvarchar(50)
);
