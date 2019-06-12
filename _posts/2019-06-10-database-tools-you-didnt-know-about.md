---
layout: article
title: Database tools you didn't know about
description: A complete list of tools for database people and developers
category: blog
---

But you already know all the tools don't you?

That's what I thought too before I started working on [SQL Schema
Explorer](http://schemaexplorer.io/). Even after 18 years working with
databases it turns out I only knew a fraction of the tools that are out there.
Below is all the things I've found and that have been shared with me as I've
been on this journey and a few I already knew. Hopefully there's more than a
few that you didn't know of and maybe you'll pick a few up and add them to your
toolbox to up your database game.

## Microsoft Sql Server

### You can now run MSSQL on open source

Did you know Microsoft SQL Server (aka mssql) is now available on both linux
natively and in docker containers? It's the real deal, not like mono vs .net

* [SQL Server on Linux](https://docs.microsoft.com/en-us/sql/linux/sql-server-linux-setup)
* [SQL Server in Docker](https://docs.microsoft.com/en-us/sql/linux/quickstart-install-connect-docker?view=sql-server-2017)

Here's all it takes to fire up mssql, the only pre-requisite is docker itself.

	docker run -e 'ACCEPT_EULA=Y' -e 'SA_PASSWORD=your_new_sa_passwod' \
	-p 1433:1433 --name mssql1 \
	-d mcr.microsoft.com/mssql/server:2017-latest

I don't know about you but one less reason to fire up the Windows VM sure does
make me happy. Combined with dotnet core I haven't fired up Windows in months
now.

### Management studio

Okay you know this one but I have to mention it.

It has awkward but functional diagram support. You can version control these
diagrams and move them between servers with
https://github.com/timabell/database-diagram-scm which is worth knowing about
if you ever use the ssms diagrams.

### SSMS Tools Pack

* https://www.ssmstoolspack.com/

## Cross-database / todo

* 2012 announce https://www.postgresql.org/about/news/1429/
* EZ data browser - http://www.softimum-solutions.com/Data-Browser/Purchase.aspx
* http://sqleo.sourceforge.net/ - https://www.youtube.com/watch?v=emDrdj0IxNI
* http://sqlfiddle.com/
* http://www.dbschema.com/sqlite-designer-tool.html
* http://www.magereverse.com/
* http://www.sqlservercentral.com/articles/Tools/157911/
* https://dataedo.com/
* https://dbvis.com/
* https://docs.microsoft.com/en-us/sql/azure-data-studio
* https://github.com/ajdeziel/SQL-Data-Viewer - abandoned
* https://github.com/preston/railroady/
* https://help.talend.com/reader/ISPDm8GQ6s0HN0348QulWg/Ij~7tBlW8im63rAGnGHT3A
* https://portableapps.com/apps/development/database_browser_portable
* https://redash.io/
* https://sequelpro.com/
* https://softwarerecs.stackexchange.com/questions/11346/tool-to-visualize-sql-database-schema
* https://sqldbm.com/en/
* https://sqlitestudio.pl/
* https://www.codediesel.com/data/5-tools-to-visualize-database-schemas/
* https://www.codediesel.com/data/5-tools-to-visualize-database-schemas/
* https://www.datasparc.com/
* https://www.dbsoftlab.com/online-tutorials/active-table-editor-online-tutorials.html
* https://www.devart.com/dbforge/sql/studio/
* https://www.idera.com/er-studio-data-architect-saftware
* https://www.idera.com/er-studio-data-architect-software
* https://www.jetbrains.com/datagrip/ - https://www.youtube.com/watch?v=Xb9K8IAdZNg
* https://www.metabase.com/
* https://www.navicat.com/en/products/navicat-data-modeler
* https://www.schemacrawler.com/
* https://www.sqlservercentral.com/articles/microsoft-sql-server-utilities-and-tools-1
* intellij/rider/rubymine etc all have the datagrip capabilities built in
* razorsql
* redgate sql prompt

## Places to find even more database tools

* https://en.wikipedia.org/wiki/Comparison_of_database_tools
* https://www.quora.com/How-do-I-generate-an-entity-relationship-diagram-for-a-SQLite-database-file?share=1
* https://www.quora.com/What-are-some-good-online-database-schema-design-tool-with-larger-days-of-expiry
* https://alternativeto.net/software/mysql-workbench/

## The end

I hope you found at least a few you didn't know about and that they make your life better in some way. Please do tell me the story of how this helped you on [email](tim@schemaexplorer.io) or [twitter](https://twitter.com/tim_abell).
Did I miss something? If you wish to improve this article please ping me a PR with additions here: https://github.com/timabell/sdv-website or just [email me](tim@schemaexplorer.io).

I'm not being paid to promote these, these are not affiliate links, I share this learning with you all for free so that we can all enjoy our work with databases more, and create better more reliable databases for ourselves, our clients and our projects.

If you want to be notified of new articles, sign up to the mailing list (which currently is also the trial download list).

Till next time!
