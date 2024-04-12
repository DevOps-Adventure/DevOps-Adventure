drop table if exists user;
create table user (
  user_id integer primary key autoincrement,
  username string not null,
  email string not null,
  pw_hash string not null
);

CREATE UNIQUE INDEX idx_username ON user(username);

-- CREATE CLUSTERED INDEX <index_name>
-- ON [schema.]<table_name>(column_name [asc|desc]);

drop table if exists follower;
create table follower (
  who_id integer,
  whom_id integer
);

drop table if exists message;
create table message (
  message_id integer primary key autoincrement,
  author_id integer not null,
  text string not null,
  pub_date integer,
  flagged integer
);
