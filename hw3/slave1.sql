CHANGE MASTER TO MASTER_HOST='mysql',
        MASTER_PORT=3306, MASTER_USER='repl',
        MASTER_PASSWORD='pass', master_auto_position =1;

start slave;
show slave status ;


stop slave;
reset slave;
reset master;

CHANGE MASTER TO MASTER_HOST='mysql',
        MASTER_PORT=3306, MASTER_USER='repl',
        MASTER_PASSWORD='pass', master_auto_position =1;
start slave;

select count(*) from otus.users;

FLUSH PRIVILEGES ;

select user, host from mysql.user;

flush privileges ;

set global server_id =3;
show variables like '%gtid%';
set global sql_slave_skip_counter  = 1;

USE otus;
DROP TABLE IF EXISTS users CASCADE;

CREATE TABLE users
(
    uuid        BINARY(16) PRIMARY KEY,
    username    VARCHAR(50) UNIQUE NOT NULL,
    first_name  VARCHAR(50)        NOT NULL,
    second_name VARCHAR(50)        NOT NULL,
    age         INT                NOT NULL,
    sex         VARCHAR(1)         NOT NULL,
    biography   TEXT               NOT NULL,
    city        VARCHAR(50)        NOT NULL,
    password    VARCHAR(255)       NOT NULL
);

show processlist;

drop schema otus;

SET @@GLOBAL.read_only = ON;

SHOW PROCESSLIST;

SET GTID_NEXT=automatic;

# SEMi SYNC
INSTALL PLUGIN rpl_semi_sync_slave SONAME 'semisync_slave.so';

SET GLOBAL rpl_semi_sync_slave_enabled = 1;

STOP SLAVE IO_THREAD;
START SLAVE IO_THREAD;

# promote to master
CREATE USER 'repl'@'%' IDENTIFIED BY '_';
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';
STOP SLAVE;
RESET MASTER;
show master status ;