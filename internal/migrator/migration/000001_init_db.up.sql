create table customer 
    (id serial primary key,
    name varchar(255) not null,
    surname varchar(255) not null,
    phone varchar(20) not null,
    masterpassword varchar(255) not null);