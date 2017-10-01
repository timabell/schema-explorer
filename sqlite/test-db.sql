-- sqlite example db

create table foo (
	id integer primary key,
	name text
);
create table woof (
	id integer primary key,
	name text,
	fooId integer,
	foreign key (fooId) references foo(id)
);
insert into foo (id, name) values
	(1, "raaa"),
	(2, "meow")
;
insert into woof (id, name, fooId) values
	(10, "muzz", 1),
	(11, "waggy", 1),
	(12, "pads", null)
;
select * from foo;
select * from woof;

CREATE TABLE "widget" (
	`code`	nvarchar(10) NOT NULL,
	`name`	TEXT NOT NULL,
	PRIMARY KEY(code)
)
CREATE TABLE "part" (
  id int not null PRIMARY key autoincrement,
	`widgetCode`	nvarchar(10) NOT NULL,
	`name`	TEXT NOT NULL,
	foreign key ([widgetCode]) references widget (code)
)

insert into widget(code, name)
values('FROB1', 'frobnizer'),
('PAN', 'Saucepan');
insert into part(widgetCode, name)
values ('FROB1', 'pinkyponk'),
('FROB1', 'ninkynonk'),
('PAN', 'handle'),
('PAN', 'pan thing');

select * from widget
inner join part on part.widgetCode = widget.code;
,