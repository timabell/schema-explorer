drop user if exists 'ssetestusr'@'localhost';
CREATE USER 'ssetestusr'@'localhost' IDENTIFIED BY 'ssetestusrpass';
GRANT ALL PRIVILEGES ON *.* TO 'ssetestusr'@'localhost' WITH GRANT OPTION;
GRANT RELOAD,PROCESS ON *.* TO 'ssetestusr'@'localhost';
FLUSH PRIVILEGES;
select user, host from mysql.user where user = 'ssetestusr';