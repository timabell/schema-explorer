---
layout: default
title: Getting started
category: help
---

# Getting Started

Thanks for giving schema explorer a try!

Please [get in touch](mailto:sse@timwise.co.uk) if you have any problems getting up and running or if you have any suggestions.

## 1. Installation

1. Use the download link from your email to download `sdv.zip`
1. *Windows only*: [Unblock the zip file](https://weblogs.asp.net/dixin/understanding-the-internet-file-blocking-and-unblocking).
	1. Right-click > properties > general > "unblock"
1. Extract the contents of the zip file to a location of your choice.
1. That's it! No need to install to the system, the program can be run from anywhere.

## 2. Running the program

### Windows

1. Open windows explorer in the extracted folder.
1. Open folder sdv
1. Open folder "windows"
1. You should see `sql-data-viewer.exe`
1. Go to the address bar and type `cmd` and press enter to open a command window in that folder
1. Type `sql-data-viewer.exe` in the command window and press enter. You should see the help text.


### Linux

1. Open a terminal window in the extracted sdv/linux/ folder
1. Run `.\sdv-linux-x64`
1. You should see the help text displayed

### Mac

1. Download the zip file
1. Unzip the zip file
1. Double-click the application file: `sse/mac/schemaexplorer`
1. You'll get a warning about untrusted developers, press ok to dismiss the warning
1. Go to system preferences in the apple menu
1. Open "Security & Privacy"
1. An "open anyway" button will have appeared for "schemaexplorer" - click it to run schema explorer

See https://www.wikihow.com/Install-Software-from-Unsigned-Developers-on-a-Mac for detailed instructions and screenshots for unblocking untrusted programs.

Note: that you should avoid doing the above as a rule, but until I see enough demand from mac users it's not worth investing in figuring out how to sign the executables to keep OSX happy.

## 3. Connecting

Windows/linux have different binary names, use the the relevant one. The arguments are the same either way.

See the command help output for more options including links to full documentation of available connection string options.

### Microsoft Sql Server

By default sql server only listens on "shared memory" which isn't supported.

Enable tcp/ip connections on your local sql server:

1. Open Sql Server Configuration Manager
1. Protocols
1. Enable tcp/ip
1. Properties
1. Enable the ip6 tcp/ip listener (`[::1]`)
1. Restart the mssql service

Connect to localhost (integrated auth):

	sql-data-viewer.exe --driver mssql --mssql-database mydatabase --mssql-host "[::1]"

To connect with explicit user (aka sql-auth) add:

	--mssql-user tim --mssql-password battery-horse

Connect to a different database server

	sql-data-viewer.exe --driver mssql --mssql-database mydatabase --mssql-host some-other-server

Connect with advanced options in connection string

	sql-data-viewer.exe --driver mssql --mssql-connection-string ""


### Microsoft Sql Server Express

Sql Express listens on named pipes by default:

	server=np:\\.\pipe\MSSQL$SQLEXPRESS\sql\query;database=mydatabase

### Sqlite

	sql-data-viewer.exe --driver sqlite --sqlite-file your/sqlitefile.db

### Postgres

Connect to localhost with default username/password on default port:

	sql-data-viewer.exe --driver pg --pg-database mydatabase

Connect with explicit user

	sql-data-viewer.exe --driver pg --pg-database mydatabase --pg-user tim --pg-password battery-horse

Connect with explicit host/port

	sql-data-viewer.exe --driver pg --pg-database mydatabase --pg-host localhost --pg-port 5432

Connect with advanced options in connection string

If you are getting "panic: pq: SSL is not enabled on the server" then you'll need sslmode=disable in a connection string as follows:

	sql-data-viewer.exe --driver pg --pg-connection-string "postgres://postgres:postgres@localhost/manage?sslmode=disable"

