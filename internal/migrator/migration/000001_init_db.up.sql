create table customer 
    (id serial primary key,
    name varchar(255) not null,
    login varchar(255) not null unique,
    masterpassword varchar(255) not null);

create table pass 
    (id serial primary key,
    pwd varchar(255) not null,
    title varchar(255) not null,
    description varchar(1000) not null,
    customer_id integer,
    foreign key (customer_id) references customer (id),
    unique (title, customer_id));

create table bfile
    (id serial primary key,
    file_name varchar(255) not null,
    title varchar(255) not null,
    description varchar(1000),
    cloud_id varchar(255) not null,
    customer_id integer,
    foreign key (customer_id) references customer (id),
    unique (title, customer_id));

create table card 
    (id serial primary key,
    num varchar(16) not null,
    expire varchar(4) not null,
    customer_id integer,
    description varchar(1000),
    cvc varchar(4),
    foreign key (customer_id) references customer (id),
    unique(num, customer_id));

create table jwttoken 
    (id serial primary key,
    customer_id integer,
    token varchar(255),
    is_valid boolean,
    foreign key (customer_id) references customer (id));