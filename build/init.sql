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