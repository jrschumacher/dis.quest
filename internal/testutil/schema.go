// Package testutil provides test database schema and utilities
package testutil

import (
	"database/sql"
)

// CreateTestSchema creates the test database schema
func CreateTestSchema(db *sql.DB) error {
	schema := `
	-- Topics table
	CREATE TABLE IF NOT EXISTS quest_dis_topic (
		did TEXT NOT NULL,
		rkey TEXT NOT NULL,
		subject TEXT NOT NULL,
		initial_message TEXT NOT NULL,
		category TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		selected_answer TEXT,
		PRIMARY KEY (did, rkey)
	);

	-- Messages table
	CREATE TABLE IF NOT EXISTS quest_dis_message (
		did TEXT NOT NULL,
		rkey TEXT NOT NULL,
		topic_did TEXT NOT NULL,
		topic_rkey TEXT NOT NULL,
		parent_message_rkey TEXT,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		PRIMARY KEY (did, rkey),
		FOREIGN KEY (topic_did, topic_rkey) REFERENCES quest_dis_topic(did, rkey)
	);

	-- Participation table
	CREATE TABLE IF NOT EXISTS quest_dis_participation (
		did TEXT NOT NULL,
		topic_did TEXT NOT NULL,
		topic_rkey TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		PRIMARY KEY (did, topic_did, topic_rkey),
		FOREIGN KEY (topic_did, topic_rkey) REFERENCES quest_dis_topic(did, rkey)
	);

	-- Indexes for better performance
	CREATE INDEX IF NOT EXISTS idx_topic_category ON quest_dis_topic(category);
	CREATE INDEX IF NOT EXISTS idx_topic_created_at ON quest_dis_topic(created_at);
	CREATE INDEX IF NOT EXISTS idx_message_topic ON quest_dis_message(topic_did, topic_rkey);
	CREATE INDEX IF NOT EXISTS idx_message_parent ON quest_dis_message(parent_message_rkey);
	CREATE INDEX IF NOT EXISTS idx_participation_user ON quest_dis_participation(did);
	CREATE INDEX IF NOT EXISTS idx_participation_topic ON quest_dis_participation(topic_did, topic_rkey);
	`

	_, err := db.Exec(schema)
	return err
}