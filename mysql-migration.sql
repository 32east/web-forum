CREATE DATABASE IF NOT EXISTS `web-forum`;

CREATE TABLE IF NOT EXISTS `users` (
	id INT PRIMARY KEY AUTO_INCREMENT,
	login TEXT NOT NULL,
	password TEXT NOT NULL,
	username TEXT NOT NULL,
	email TEXT NOT NULL,
	avatar TEXT,
	description TEXT,
    sign_text TEXT,
	created_at DATETIME NOT NULL,
	updated_at DATETIME
);

CREATE TABLE IF NOT EXISTS `forums` (
    id INT PRIMARY KEY,
    forum_name TEXT,
    forum_description TEXT
);

CREATE TABLE IF NOT EXISTS `topics` (
    id INT PRIMARY KEY AUTO_INCREMENT,
    forum_id INT NOT NULL,
    topic_name TEXT NOT NULL,
    topic_message TEXT NOT NULL,
    created_by INT NOT NULL,
    create_time DATETIME NOT NULL,
    update_time DATETIME,
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (forum_id) REFERENCES forums(id)
);

CREATE TABLE IF NOT EXISTS `messages` (
    id INT PRIMARY KEY AUTO_INCREMENT,
    topic_id INT NOT NULL,
    account_id INT NOT NULL,
    message TEXT NOT NULL,
    create_time DATETIME NOT NULL,
    update_time DATETIME,
    FOREIGN KEY (topic_id) REFERENCES topics(id),
    FOREIGN KEY (account_id) REFERENCES users(id)
);

CREATE INDEX message_index ON messages (topic_id) USING BTREE;
CREATE INDEX topics_index ON topics (id) USING BTREE;
CREATE INDEX users_index ON users (id) USING BTREE;