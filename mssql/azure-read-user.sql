-- from http://sqlblog.com/blogs/jamie_thomson/archive/2012/03/27/adventureworks2012-now-available-to-all-on-sql-azure.aspx

-- run this on master
create login sseRO with password='Startups 4 the rest of us';

-- reconnect to different db. USE doesn't work with azure

-- run this in the database
CREATE USER sseRO FOR LOGIN sseRO;
--GRANT VIEW DEFINITION ON Database::AdventureWorksLT TO sseROrole;
--GRANT VIEW DATABASE STATE ON Database::AdventureWorksLT TO sseROrole;
--GRANT SHOWPLAN TO sseROrole;
EXEC sp_addrolemember 'db_datareader','sseRO';

--http://stackoverflow.com/a/31447248/10245
select m.name as Member, r.name as Role
from sys.database_role_members
inner join sys.database_principals m on sys.database_role_members.member_principal_id = m.principal_id
inner join sys.database_principals r on sys.database_role_members.role_principal_id = r.principal_id
