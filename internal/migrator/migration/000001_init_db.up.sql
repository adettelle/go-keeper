create table customer 
    (id serial primary key,
    name varchar(255) not null,
    email varchar(255) not null unique,
    masterpassword varchar(255) not null);