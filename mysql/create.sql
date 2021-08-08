DROP DATABASE entry_task;

CREATE DATABASE entry_task DEFAULT CHARSET = utf8mb4 DEFAULT COLLATE = utf8mb4_unicode_ci;

USE entry_task;

CREATE TABLE user (
    username VARCHAR(100) PRIMARY KEY NOT NULL,
    nickname VARCHAR(100) NOT NULL,
    password VARCHAR(250) NOT NULL
) DEFAULT CHARSET=utf8mb4 COLLATE utf8mb4_unicode_ci;

