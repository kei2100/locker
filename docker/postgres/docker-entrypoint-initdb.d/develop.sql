CREATE DATABASE develop;
CREATE ROLE develop WITH LOGIN PASSWORD 'develop';
\c develop
CREATE SCHEMA develop AUTHORIZATION develop;
