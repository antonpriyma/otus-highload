show master status ;

RESET MASTER;
select count(*) from otus.users;

select user,host from mysql.user;

CREATE DATABASE newdb13;

show global variables like 'gtid_executed';
show variables like  '%binlog%';

SET @@GLOBAL.read_only = ON;

SHOW PROCESSLIST;

select * from mysql.user;

set global enforce_gtid_consistency = "ON";

set global log_slave_updates ="ON";

SHOW STATUS LIKE 'ONGOING_ANONYMOUS_TRANSACTION_COUNT';

select * from mysql.gtid_executed;

FLUSH LOGS;

show global variables like 'gtid_purged';


# SEMI SYNC
INSTALL PLUGIN rpl_semi_sync_master SONAME 'semisync_master.so';

SELECT PLUGIN_NAME, PLUGIN_STATUS
FROM INFORMATION_SCHEMA.PLUGINS
WHERE PLUGIN_NAME LIKE '%semi%';

SET GLOBAL rpl_semi_sync_master_enabled = 1;