-- todo: This file fails under jtds. Needs to be run in windows tools. figure out why and fix it.
-- just get connection reset with azure sql. Odd.


-- Permanent copy of schema documentation.
-- In a format suitable for easy source control and hand-editing.

-- At the bottom you'll find a commented-out `select` for generating the
--   insert block from an existing schema's extended properties.

-- This script will add/update/remove descriptions from a the schema it's
--   run against to bring them into line with the below description list.

-- https://gist.github.com/timabell/6fbd85431925b5724d2f

set nocount on;
set xact_abort on;

begin tran
declare @doc table (id int primary key identity(1,1), [table] sysname, [column] sysname null, [description] sql_variant);

insert into @doc ([table], [column], [description]) values
    ('person', null, 'somebody to love'), -- table description
    ('person', 'personName', 'say my name!')
;

declare @action_delete varchar(10) = 'delete';
declare @action_update varchar(10) = 'update';
declare @action_add varchar(10) = 'add';

--select * from @doc;
/*
-- see what's changed
select
    isnull(doc.[table], existing.[table]) [table],
    isnull(doc.[column], existing.[column]) [column],
    existing.description existingDescription,
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
            tbl.name [table],
            col.name [column],
            ep.value [description]
        from sys.extended_properties ep
            inner join sys.objects tbl on tbl.object_id = ep.major_id
                and tbl.name not like '\_\_%' escape '\' -- ignore the ready-roll object(s)
            left outer join sys.columns col on col.object_id = ep.major_id and col.column_id = ep.minor_id
        where ep.name = 'MS_Description'
    ) existing
        on existing.[table] = doc.[table]
            and (existing.[column] = doc.[column] or (existing.[column] is null and doc.[column] is null))
where doc.description <> existing.description or doc.description is null or existing.description is null
*/

-- set up list of properties to check
declare mergeList cursor for
    select
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
                tbl.name [table],
                col.name [column],
                ep.value [description]
            from sys.extended_properties ep
                inner join sys.objects tbl on tbl.object_id = ep.major_id
                    and tbl.name not like '\_\_%' escape '\' -- ignore the ready-roll object(s)
                left outer join sys.columns col on col.object_id = ep.major_id and col.column_id = ep.minor_id
            where ep.name = 'MS_Description'
        ) existing
            on existing.[table] = doc.[table]
                and (existing.[column] = doc.[column] or (existing.[column] is null and doc.[column] is null))
    where doc.description <> existing.description or doc.description is null or existing.description is null
;

open mergeList;

declare @table sysname;
declare @column sysname;
declare @newDescription sql_variant;
declare @action varchar(10);

fetch next from mergeList into @table, @column, @newDescription, @action;
while @@FETCH_STATUS = 0
begin
    --print concat(@table, '.', @column, ' - ', @action);
    if @action = @action_add
    begin
        if @column is null
        begin
            print 'adding description for ' + @table;
            exec sys.sp_addextendedproperty @name=N'MS_Description', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE',
                @level1name=@table, @value=@newDescription
        end
        else
        begin
            print 'adding description for ' + @table + '.' + @column;
            exec sys.sp_addextendedproperty @name=N'MS_Description', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level2type=N'COLUMN',
                @level1name=@table, @level2name=@column, @value=@newDescription
        end
    end
    else if @action = @action_update
    begin
        if @column is null
        begin
            print 'updating description for ' + @table;
            exec sys.sp_updateextendedproperty @name=N'MS_Description', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE',
                @level1name=@table, @value=@newDescription
        end
        else
        begin
            print 'updating description for ' + @table + '.' + @column;
            exec sys.sp_updateextendedproperty @name=N'MS_Description', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level2type=N'COLUMN',
                @level1name=@table, @level2name=@column, @value=@newDescription
        end
    end
    else if @action = @action_delete
    begin
        if @column is null
        begin
            print 'dropping description for ' + @table;
            exec sys.sp_dropextendedproperty @name=N'MS_Description', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE',
                @level1name=@table
        end
        else
        begin
            print 'dropping description for ' + @table + '.' + @column;
            exec sys.sp_dropextendedproperty @name=N'MS_Description', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level2type=N'COLUMN',
                @level1name=@table, @level2name=@column
        end
    end
    fetch next from mergeList into @table, @column, @newDescription, @action;
end
close mergeList;
deallocate mergeList;

/*
-- see existing extended props:
select
    tbl.name [table],
    col.name [column],
    ep.value [description]
from sys.extended_properties ep
    inner join sys.objects tbl on tbl.object_id = ep.major_id
    left outer join sys.columns col on col.object_id = ep.major_id and col.column_id = ep.minor_id
where ep.name = 'MS_Description'
order by tbl.name, ep.minor_id

-- output in a format suitable for inclusion in this script as a reference copy for source-control
-- run it, copy paste the result into the above insert into @doc, swap the last comma for a semi-colon.
-- don't use results-as-text as it truncates long descriptions.
select
    '(''' + 
        tbl.name +
    ''', ' + 
        iif(col.name is null, 'null', '''' + col.name + '''') +
    ', N''' + 
        replace(cast(ep.value as nvarchar(max)), '''', '''''') +
    '''),'
from sys.extended_properties ep
    inner join sys.objects tbl on tbl.object_id = ep.major_id
    left outer join sys.columns col on col.object_id = ep.major_id and col.column_id = ep.minor_id
where ep.name = 'MS_Description'
order by tbl.name, ep.minor_id
*/
commit
