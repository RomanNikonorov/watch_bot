create table servers
(
id   bigserial
constraint servers_pk
primary key,
name text not null,
url  text not null
);