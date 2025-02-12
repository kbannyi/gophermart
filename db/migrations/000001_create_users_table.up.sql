CREATE TABLE users
(
    id       varchar(100) PRIMARY KEY,
    login    varchar(100) UNIQUE,
    password varchar(100)
);