/*

# About

    Source: https://gist.github.com/timabell/6fbd85431925b5724d2f

    Permanent copy of schema documentation in a format suitable for
    easy source control and hand-editing.

    This script template was created by Tim Abell, (c) 2015-2018
    MIT Licence.
    The @doc data herein is for your own project, and is not necessarily MIT licenced.
    See containing project's docs for more info.

    WARNING:
    Take a backup before running this script.
    When run this script will remove / update all ms_description attributes that don't match the data below without further warning.

# Usage

    1) Add this file verbatim to your source control,

    2) Run the commented out statement in Section 1 to see the current state of your descriptions (or use http://schemaexplorer.io/ )

    3) Run the statement in Section 2 to generate the the source-controllable statements

    4) Copy-paste those from the SSMS / Visual Studio output into the @doc INSERT block in Section 3.

       ! Don't use "results-as-text" as it truncates long descriptions. Stick to grid view, and even then watch out for truncation. !

    5) Check in the updated documentation.

## You now have two options depending on how you like to work for making changes to your schema documentation ##

    a) Manually edit the below @doc table. Then run this whole file against your db to apply your changes.

    or

    b) Use a tool like RedGate SQL Doc or SSMS to update the descriptions on a running sql server, and then steps 3-5 above to source-control your changes.

# Related things you should consider

    Throw away any manual source controlled add / update / delete statements you had for ms_description.

    ReadyRoll by RedGate is great for migrations, and has support for scripts to be run every time, make this one of those.

---

Read the rest of the comments for more usage info.
*/


set nocount on;
set xact_abort on;

begin tran
declare @doc table (id int primary key identity(1,1), [schema] sysname, [table] sysname, [column] sysname null, [description] sql_variant);

/* Section 1 **********

-- see existing extended props:
select sch.name [schema], tbl.name [table], col.name [column], ep.value [description]
from sys.extended_properties ep
    inner join sys.objects tbl on tbl.object_id = ep.major_id
    inner join sys.schemas sch on sch.schema_id = tbl.schema_id
    left outer join sys.columns col on col.object_id = ep.major_id and col.column_id = ep.minor_id
where ep.name = 'MS_Description' order by tbl.name, ep.minor_id

*/

/* Section 2 **********
-- Output existing ms_description attributes in a format suitable for updating the @doc data below.
-- i.e. use this statement to go from a real database to a source controlled file
-- run this statement then copy-paste the result into the below @doc INSERT statement, swap the last comma for a semi-colon.
-- don't use results-as-text as it truncates long descriptions.

select
    '(''' +
    sch.name +
    ''', ''' +
        tbl.name +
    ''', ' +
        iif(col.name is null, 'null', '''' + col.name + '''') +
    ', N''' +
        replace(cast(ep.value as nvarchar(max)), '''', '''''') +
    '''),'
from sys.extended_properties ep
    inner join sys.objects tbl on tbl.object_id = ep.major_id
    inner join sys.schemas sch on sch.schema_id = tbl.schema_id
    left outer join sys.columns col on col.object_id = ep.major_id and col.column_id = ep.minor_id
where ep.name = 'MS_Description'
order by tbl.name, ep.minor_id

*/

/* Section 3 **********/
insert into @doc
    ([schema], [table], [column],       [description]) values
    -- edit below here manually this to update your ms_descriptions,
    ('dbo',    'person', null,         'somebody to love'), -- table description (column is null)
    ('dbo',    'person', 'personName', 'say my name!'),
    ('kitchen', 'sink', null, 'call a plumber!!!'),
    ('kitchen', 'sink', 'sinkId', 'gotta number your sinks man!')
    -- /end editable section
;

declare @action_delete varchar(10) = 'delete';
declare @action_update varchar(10) = 'update';
declare @action_add varchar(10) = 'add';

--select * from @doc;
/*
-- see what's changed (exact copy of merge cursor sql below)
select
    isnull(doc.[schema], existing.[schema]) [schema],
    isnull(doc.[table], existing.[table]) [table],
    isnull(doc.[column], existing.[column]) [column],
    doc.description newDescription,
    case
        when doc.id is null then @action_delete
        when existing.[table] is null then @action_add
        else @action_update
    end as action
from
    @doc doc
    full outer join
    (
        select
            sch.name [schema],
            tbl.name [table],
            col.name [column],
            ep.value [description]
        from sys.extended_properties ep
            inner join sys.objects tbl on tbl.object_id = ep.major_id
                --and tbl.name not like '\_\_%' escape '\' -- ignore the ready-roll object(s)
            inner join sys.schemas sch on sch.schema_id = tbl.schema_id
            left outer join sys.columns col on col.object_id = ep.major_id and col.column_id = ep.minor_id
        where ep.name = 'MS_Description'
    ) existing
        on existing.[schema] = doc.[schema] and existing.[table] = doc.[table]
            and (existing.[column] = doc.[column] or (existing.[column] is null and doc.[column] is null))
where doc.description <> existing.description or doc.description is null or existing.description is null
*/
/* you can run between begin-tran and here to preview changes that will be made */

-- set up list of properties to check
declare mergeList cursor for
    select
        isnull(doc.[schema], existing.[schema]) [schema],
        isnull(doc.[table], existing.[table]) [table],
        isnull(doc.[column], existing.[column]) [column],
        doc.description newDescription,
        case
            when doc.id is null then @action_delete
            when existing.[table] is null then @action_add
            else @action_update
        end as action
    from
        @doc doc
        full outer join
        (
            select
                sch.name [schema],
                tbl.name [table],
                col.name [column],
                ep.value [description]
            from sys.extended_properties ep
                inner join sys.objects tbl on tbl.object_id = ep.major_id
                    --and tbl.name not like '\_\_%' escape '\' -- ignore the ready-roll object(s)
                inner join sys.schemas sch on sch.schema_id = tbl.schema_id
                left outer join sys.columns col on col.object_id = ep.major_id and col.column_id = ep.minor_id
            where ep.name = 'MS_Description'
        ) existing
            on existing.[schema] = doc.[schema] and existing.[table] = doc.[table]
                and (existing.[column] = doc.[column] or (existing.[column] is null and doc.[column] is null))
    where doc.description <> existing.description or doc.description is null or existing.description is null
;

open mergeList;

declare @schema sysname;
declare @table sysname;
declare @column sysname;
declare @newDescription sql_variant;
declare @action varchar(10);

fetch next from mergeList into @schema, @table, @column, @newDescription, @action;
while @@FETCH_STATUS = 0
begin
--     print concat(@schema, '.', @table, '.', @column, ' - ', @action);
    if @action = @action_add
    begin
        if @column is null
        begin
--             print 'adding description for ' + @schema + '.' + @table;
            exec sys.sp_addextendedproperty @name=N'MS_Description', @level0type=N'SCHEMA', @level1type=N'TABLE',
                @level0name=@schema, @level1name=@table, @value=@newDescription
        end
        else
        begin
--             print 'adding description for ' + @schema + '.' + @table + '.' + @column;
            exec sys.sp_addextendedproperty @name=N'MS_Description', @level0type=N'SCHEMA', @level1type=N'TABLE', @level2type=N'COLUMN',
                @level0name=@schema, @level1name=@table, @level2name=@column, @value=@newDescription
        end
    end
    else if @action = @action_update
    begin
        if @column is null
        begin
            print 'updating description for ' + @schema + '.' + @table;
            exec sys.sp_updateextendedproperty @name=N'MS_Description', @level0type=N'SCHEMA', @level1type=N'TABLE',
                @level0name=@schema, @level1name=@table, @value=@newDescription
        end
        else
        begin
            print 'updating description for ' + @schema + '.' + @table + '.' + @column;
            exec sys.sp_updateextendedproperty @name=N'MS_Description', @level0type=N'SCHEMA', @level1type=N'TABLE', @level2type=N'COLUMN',
                @level0name=@schema, @level1name=@table, @level2name=@column, @value=@newDescription
        end
    end
    else if @action = @action_delete
    begin
        if @column is null
        begin
            print 'dropping description for ' + @schema + '.' + @table;
            exec sys.sp_dropextendedproperty @name=N'MS_Description', @level0type=N'SCHEMA', @level1type=N'TABLE',
                @level0name=@schema, @level1name=@table
        end
        else
        begin
            print 'dropping description for ' + @schema + '.' + @table + '.' + @column;
            exec sys.sp_dropextendedproperty @name=N'MS_Description', @level0type=N'SCHEMA', @level1type=N'TABLE', @level2type=N'COLUMN',
                @level0name=@schema, @level1name=@table, @level2name=@column
        end
    end
    fetch next from mergeList into @schema, @table, @column, @newDescription, @action;
end
close mergeList;
deallocate mergeList;

commit
