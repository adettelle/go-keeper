create table customer 
    (id serial primary key,
    name varchar(255) not null,
    email varchar(255) not null unique,
    masterpassword varchar(255) not null);

create table pass 
    (id serial primary key,
    pwd varchar(255) not null,
    title varchar(255) not null,
    description varchar(1000) not null,
    customer_id integer,
    unique (title, customer_id));