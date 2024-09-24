CREATE TABLE IF NOT EXISTS users (
	id serial PRIMARY KEY,
	login VARCHAR(64) NOT NULL,
	password VARCHAR(64) NOT NULL,
	username VARCHAR(64) NOT NULL,
	email VARCHAR(128) NOT NULL,
    is_admin BOOLEAN NOT NULL default false,
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
    topics_count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS topics (
    id serial PRIMARY KEY,
    forum_id INTEGER NOT NULL,
    topic_name VARCHAR(384) NOT NULL,
    created_by INTEGER NOT NULL,
    create_time TIMESTAMP NOT NULL,
    update_time TIMESTAMP,
    message_count INTEGER NOT NULL DEFAULT 1,
    parent_id integer,
    FOREIGN KEY (created_by) REFERENCES users(id) on delete cascade,
    FOREIGN KEY (forum_id) REFERENCES forums(id) on delete cascade
);

CREATE TABLE IF NOT EXISTS messages (
    id serial PRIMARY KEY,
    topic_id INTEGER NOT NULL,
    account_id INTEGER NOT NULL,
    message TEXT NOT NULL,
    create_time TIMESTAMP NOT NULL,
    update_time TIMESTAMP,
    FOREIGN KEY (topic_id) REFERENCES topics(id) on delete cascade,
    FOREIGN KEY (account_id) REFERENCES users(id) on delete cascade
);

create table if not exists tokens (
    id serial primary key,
    account_id integer not null,
    refresh_token varchar(255) not null,
    expiresAt timestamp not null,
    FOREIGN KEY (account_id) REFERENCES users(id) on delete cascade
);

create index if not exists refresh_token_index on tokens(refresh_token);
create index if not exists messages_id on messages(id);
create index if not exists messages_topic_id on messages(topic_id);
create index if not exists messages_create_time on messages(create_time);
create index if not exists messages_account_id on messages(account_id);

create index if not exists topics_id on topics(id);
create index if not exists topics_parent_id on topics(parent_id);
create index if not exists users_id on users(id);

alter table topics add FOREIGN KEY (parent_id) REFERENCES messages(id) on delete cascade;
