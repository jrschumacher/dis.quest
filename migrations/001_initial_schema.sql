-- Initial schema for dis.quest based on ATProtocol lexicons
-- quest.dis.topic, quest.dis.message, quest.dis.participation

-- Topics table - represents discussion topics
CREATE TABLE quest_dis_topic (
    did TEXT NOT NULL,
    rkey TEXT NOT NULL,
    subject TEXT NOT NULL,
    initial_message TEXT NOT NULL,
    category TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    selected_answer TEXT, -- For Q&A topics
    PRIMARY KEY (did, rkey)
);

-- Messages table - represents messages within topics
CREATE TABLE quest_dis_message (
    did TEXT NOT NULL,
    rkey TEXT NOT NULL,
    topic_did TEXT NOT NULL,
    topic_rkey TEXT NOT NULL,
    parent_message_rkey TEXT, -- NULL for top-level messages
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (did, rkey),
    FOREIGN KEY (topic_did, topic_rkey) REFERENCES quest_dis_topic(did, rkey) ON DELETE CASCADE
);

-- Participation table - tracks user participation in topics
CREATE TABLE quest_dis_participation (
    did TEXT NOT NULL,
    topic_did TEXT NOT NULL,
    topic_rkey TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'following', -- following, muted, etc.
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (did, topic_did, topic_rkey),
    FOREIGN KEY (topic_did, topic_rkey) REFERENCES quest_dis_topic(did, rkey) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_quest_dis_topic_category ON quest_dis_topic(category);
CREATE INDEX idx_quest_dis_topic_created_at ON quest_dis_topic(created_at);
CREATE INDEX idx_quest_dis_message_topic ON quest_dis_message(topic_did, topic_rkey);
CREATE INDEX idx_quest_dis_message_parent ON quest_dis_message(topic_did, topic_rkey, parent_message_rkey);
CREATE INDEX idx_quest_dis_message_created_at ON quest_dis_message(created_at);
CREATE INDEX idx_quest_dis_participation_user ON quest_dis_participation(did);
CREATE INDEX idx_quest_dis_participation_topic ON quest_dis_participation(topic_did, topic_rkey);

---- create above / drop below ----

-- Drop tables in reverse order due to foreign key constraints
DROP INDEX IF EXISTS idx_quest_dis_participation_topic;
DROP INDEX IF EXISTS idx_quest_dis_participation_user;
DROP INDEX IF EXISTS idx_quest_dis_message_created_at;
DROP INDEX IF EXISTS idx_quest_dis_message_parent;
DROP INDEX IF EXISTS idx_quest_dis_message_topic;
DROP INDEX IF EXISTS idx_quest_dis_topic_created_at;
DROP INDEX IF EXISTS idx_quest_dis_topic_category;

DROP TABLE IF EXISTS quest_dis_participation;
DROP TABLE IF EXISTS quest_dis_message;
DROP TABLE IF EXISTS quest_dis_topic;