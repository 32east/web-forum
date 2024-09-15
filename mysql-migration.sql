CREATE DATABASE IF NOT EXISTS web-forum;

CREATE TABLE IF NOT EXISTS users (
	id serial PRIMARY KEY,
	login VARCHAR(64) NOT NULL,
	password VARCHAR(64) NOT NULL,
	username VARCHAR(64) NOT NULL,
	email VARCHAR(128) NOT NULL,
    is_admin BOOLEAN NOT NULL,
    sex CHAR(1),
    avatar VARCHAR(1024),
	description VARCHAR(1024),
    sign_text VARCHAR(1024),
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS forums (
    id serial PRIMARY KEY,
    forum_name VARCHAR(128),
    forum_description VARCHAR(2048),
    topics_count INTEGER NOT NULL DEFAULT 0,
);

CREATE TABLE IF NOT EXISTS topics (
    id serial PRIMARY KEY,
    forum_id INTEGER NOT NULL,
    topic_name VARCHAR(384) NOT NULL,
    created_by INTEGER NOT NULL,
    create_time TIMESTAMP NOT NULL,
    update_time TIMESTAMP,
    message_count INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (forum_id) REFERENCES forums(id)
);

CREATE TABLE IF NOT EXISTS messages (
    id serial PRIMARY KEY,
    topic_id INTEGER NOT NULL,
    account_id INTEGER NOT NULL,
    message TEXT NOT NULL,
    create_time TIMESTAMP NOT NULL,
    update_time TIMESTAMP,
    FOREIGN KEY (topic_id) REFERENCES topics(id),
    FOREIGN KEY (account_id) REFERENCES users(id)
);

create index messages_id on messages(id);
create index messages_topic_id on messages(topic_id);
create index messages_create_time on messages(create_time);
create index messages_account_id on messages(account_id);

create index topics_id on topics(id);
create index users_id on users(id);