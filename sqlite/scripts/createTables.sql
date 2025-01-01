-- User Table
CREATE TABLE IF NOT EXISTS user (
	id INTEGER PRIMARY KEY,
	username STRING NOT NULL UNIQUE,
	password STRING NOT NULL
);

-- Bucket Table
CREATE TABLE IF NOT EXISTS bucket (
	id INTEGER PRIMARY KEY,
	name STRING NOT NULL,
	user_id INTEGER NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users (id)
);

-- Transaction Table
CREATE TABLE IF NOT EXISTS transaction_item (
	id INTEGER PRIMARY KEY,
	name STRING NOT NULL,
	month INTEGER NOT NULL,
	year INTEGER NOT NULL,
	price REAL NOT NULL,
	user_id INTEGER NOT NULL,
	bucket_id INTEGER NOT NULL,
	FOREIGN KEY (user_id) REFERENCES user (id)
	FOREIGN KEY (bucket_id) REFERENCES bucket (id)
);

