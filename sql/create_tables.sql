-- Creation of product table
CREATE TABLE IF NOT EXISTS users (
  username varchar(250) NOT NULL,
  password varchar(250) NOT NULL
);

-- Filling of products
INSERT INTO users VALUES('test-user', '42b27efc1480b4fe6d7eaa5eec47424d');
INSERT INTO users VALUES('flag-user', 'FLAG{{AI_will_take_over_the_world_RIP!!}}');
