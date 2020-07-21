---
layout: article
title: 5 ways to make your database better - by Tim Abell
description: Step up your database game with these top tips.
category: blog
---

## [1] Documentation

Shoot me okay, but maintenance of software is [insert large number here] times
the cost of creation, especially with relational databases. You are a pro
working for a client, you owe it to them to make it possible for them to have
future staff (and yourself!) be as effective as possible. You put all that
effort into figuring out why a column should exist and have that name, now
share that knowledge before you move on to the next greenfield project

* [Redgate SqlDoc](https://www.red-gate.com/products/sql-development/sql-doc/)
	is great for rapidly adding missing documentation.
* [SchemaSpy](http://schemaspy.org/) generates static html sites making it easy
	to see what documentation there is (or isn't!) and share it with the team.
	It's free & open source (although a bit fiddly to set up and run). It has
	particularly nice clickable diagrams.
* [Dataedo generates static html sites &
	pdfs](https://dataedo.com/tutorials/getting-started/generating-database-documentation)
	as well, and is commercial and slicker than SchemaSpy
* This [gist for source-controlling ms_description
	attributes](https://gist.github.com/timabell/6fbd85431925b5724d2f) gives you
	a two-way source-controllable / editable list of your documentation in SQL
	Server
* [SQL Schema Explorer](https://timabell.github.io/schema-explorer/) generates dynamic html sites
	making it easy to see what documentation there is and share it with the team.

## [2] Refactor your database

Migrations are a thing now. Use them. You refactor your code, why wouldn't you
refactor your database? Stop leaving landmines for future people - misleading
names, bad structures etc. Use the redgate tools (ready-roll etc), use your
orm’s tools (EF migrations, active record migrations). Yes you have to deal
with data, but it’s the exception not the rule that it’s going to take hours to
run because of data volumes.

* [Redgate's SQL Change
	Automation](https://www.red-gate.com/products/sql-development/sql-change-automation/)
	(formerly
	[ReadyRoll](https://www.red-gate.com/blog/working/from-release-engineer-to-readyroll-founder-and-redgate-product-manager))
	is an opinionated tool for creating and running database migrations, it even
	generates Database Administrator (DBA) friendly pure-sql deployment packages.
	Very impressive!
* [Redgate's SQL Source Control supports
	migrations](https://documentation.red-gate.com/soc6/common-tasks/working-with-migration-scripts)
* I've been using [EF Core
	migrations](https://docs.microsoft.com/en-us/ef/core/managing-schemas/migrations/)
	recently and they work well. There are equivalents for all the major
	platforms.

## [3] Enforce data integrity

Does your app fall over if the data is bad? Databases have many powerful ways
of enforcing the rules your code relies on: nullability, foreign keys, [check
constraints](https://www.w3schools.com/SQL/sql_check.asp), unique constraints.
Stop the bad data before it even gets in there. Now your database is enforcing
these rules your code doesn't have to handle violations of them when reading
data because they'll never happen

## [4] Integration testing

You have an ORM. Great. You have unit tests. Great. But where the rubber hits
the road and your code sends SQL to a real database it breaks at runtime more
often than you’d like to admit because the generated sql didn't jive with the
real database structure or data in some obscure fashion. Automate the
creation/test/destruction of your db and run full end to end integration tests.
I suggest automating from the layer below the UI to keep the tests fast. There
are many techniques for keeping the tests quick but still realistic: do end to
end smoke tests instead of individual pieces, use an in-memory database, use
[database
snapshots](https://gist.github.com/timabell/3164291#file-create-snapshot-sql)
or the fancy [sql-clone](https://www.red-gate.com/products/dba/sql-clone/index)
tool from Redgate to make creation / rollback virtually instant. Can you pull
realistic (anonymised) data from production? Better still, now you’ll catch a
whole new class of bugs before they hit prod.

* Here's [a guide from Redgate detailing one way to do continuous integration
	testing with
	databases](https://www.red-gate.com/simple-talk/sql/sql-tools/continuous-integration-for-databases-using-red-gate-tools/)

## [5] Make it visible

Are the only people that can see the database structures the coders and DBAs?
do the business owners, support people, Quality Assurance (QA) people find it a
mystery? You should be just as proud of your database as you are of your code,
by shining a light on this dark corner of your digital estate you can make it
as good as it should be, not an embarrassing backwater. By sharing the database
in an accessible form to the non-coders in your team you can help them be more
effective in their jobs.

* The html generated by [SchemaSpy](http://schemaspy.org/) can be shared on any
	webserver to let your whole team see your schema structures
* [SQL Schema Explorer](https://timabell.github.io/schema-explorer/) can be run on your network
	or cloud hosting ([schema explorer is
	dockerized](https://hub.docker.com/r/timabell/sdv/)!) to give your team easy
	access to both the schema and data within the database.

Combine these tools with a continuous integration system and you have easy
access to the bleeding edge of your databases development.

## Take action now!

1.Make a start on at least one of these improvements today.
2.Share this article with your team - get everyone motivated to improve.
3.Share this article on social media - help spread the word that our
	databases deserve better!

I hope this has inspired you to make an improvement in the often unloved
underbelly of your applications.

What do you think needs improving in the way we deal with databases? What
change did you make because of this? [Let me
know!](mailto:articles@timwise.co.uk?subject=making better databases)

Originally posted at
[http://schemaexplorer.io/blog/2018/07/10/5-ways-to-make-your-database-better.html](http://schemaexplorer.io/blog/2018/07/10/5-ways-to-make-your-database-better.html)
