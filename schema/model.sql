drop table if exists users;
create table users (
    id int(11) not null auto_increment,
    first_name varchar(255),
    last_name varchar(255),
    primary key (id)
);

drop table if exists user_groups;
create table user_groups (
    id int(11) not null auto_increment,
    name varchar(255) not null,
    primary key (id)
);

drop table if exists group_users;
create table group_users (
    id int(11) not null auto_increment,
    user_id int(11) not null,
    group_id int(11) not null,
    primary key (id),
    foreign key (user_id) references users(id),
    foreign key (group_id) references user_groups(id)
);
create index group_users_user_id on group_users(user_id);
create index group_users_group_id on group_users(group_id);

drop table if exists fields;
create table fields (
    id int not null auto_increment,
    tinyint_field tinyint(4) not null,
    tinyint_unsigned_field tinyint(4) unsigned not null,
    tinyint_nullable_field tinyint(4),
    tinyint_unsigned_nullable_field tinyint(4) unsigned,
    smallint_field smallint(6) not null,
    smallint_unsigned_field smallint(6) unsigned not null,
    smallint_nullable_field smallint(6) ,
    smallint_unsigned_nullable_field smallint(6) unsigned,
    mediumint_field mediumint(6) not null,
    mediumint_unsigned_field mediumint(6) unsigned not null,
    mediumint_nullable_field mediumint(6) ,
    mediumint_unsigned_nullable_field mediumint(6) unsigned,
    int_field int(11) not null,
    int_unsigned_field int(11) unsigned not null,
    int_nullable_field int(11) ,
    int_unsigned_nullable_field int(11) unsigned,
    bigint_field bigint(20) not null,
    bigint_unsigned_field bigint(20) unsigned not null,
    bigint_nullable_field bigint(20) ,
    bigint_unsigned_nullable_field bigint(20) unsigned,
    float_field float not null,
    float_null_field float,
    double_field double not null,
    double_null_field double ,
    tinytext_field tinytext not null,
    tinytext_null_field tinytext,
    mediumtext_field mediumtext not null,
    mediumtext_null_field mediumtext,
    text_field text not null,
    text_null_field text,
    longtext_field longtext not null,
    longtext_null_field longtext,
    varchar_filed_field varchar(255) not null,
    varchar_null_field varchar(255),
    char_filed_field char(10) not null,
    char_filed_null_field char(10),
    date_field date not null,
    date_null_field date,
    datetime_field datetime not null,
    datetime_null_field datetime,
    time_field time not null,
    time_null_field time,
    timestamp_field timestamp not null,
    timestamp_null_field timestamp null,
    tinyblob_field tinyblob not null,
    tinyblob_null_field tinyblob,
    mediumblob_field mediumblob not null,
    mediumblob_null_field mediumblob,
    blob_field blob not null,
    blob_null_field blob,
    longblob_field longblob not null,
    longblob_null_field longblob,
    json_field json not null,
    json_null_field json,
    primary key (id)
);