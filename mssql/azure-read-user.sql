-- from http://sqlblog.com/blogs/jamie_thomson/archive/2012/03/27/adventureworks2012-now-available-to-all-on-sql-azure.aspx

-- run this on master
create login sdvRO with password='Startups 4 the rest of us';

-- reconnect to different db. USE doesn't work with azure

-- run this in the database
CREATE USER sdvRO FOR LOGIN sdvRO;
--GRANT VIEW DEFINITION ON Database::AdventureWorksLT TO sdvROrole;
--GRANT VIEW DATABASE STATE ON Database::AdventureWorksLT TO sdvROrole;
--GRANT SHOWPLAN TO sdvROrole;
EXEC sp_addrolemember 'db_datareader','sdvRO';

--http://stackoverflow.com/a/31447248/10245
select m.name as Member, r.name as Role
from sys.database_role_members
inner join sys.database_principals m on sys.database_role_members.member_principal_id = m.principal_id
inner join sys.database_principals r on sys.database_role_members.role_principal_id = r.principal_id
