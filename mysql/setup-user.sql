drop user if exists 'ssetestusr'@'%';
CREATE USER 'ssetestusr'@'%' IDENTIFIED BY 'ssetestusrpass';
GRANT ALL PRIVILEGES ON *.* TO 'ssetestusr'@'%' WITH GRANT OPTION;
GRANT RELOAD,PROCESS ON *.* TO 'ssetestusr'@'%';
FLUSH PRIVILEGES;
select user, host from mysql.user where user = 'ssetestusr';
