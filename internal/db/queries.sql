-- queries.sql - Central SQL query file for dis.quest
-- All SQL queries should be added to this file as documented in CLAUDE.md

-- Topics queries
-- name: CreateTopic :one
INSERT INTO quest_dis_topic (
    did, rkey, subject, initial_message, category, created_at, updated_at, selected_answer
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetTopic :one
SELECT * FROM quest_dis_topic
WHERE did = ? AND rkey = ?;

-- name: GetTopicsByCategory :many
SELECT * FROM quest_dis_topic
WHERE category = ?
ORDER BY created_at DESC
LIMIT ?;

-- name: ListTopics :many
SELECT * FROM quest_dis_topic
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateTopicSelectedAnswer :exec
UPDATE quest_dis_topic
SET selected_answer = ?, updated_at = ?
WHERE did = ? AND rkey = ?;

-- name: DeleteTopic :exec
DELETE FROM quest_dis_topic
WHERE did = ? AND rkey = ?;

-- Messages queries
-- name: CreateMessage :one
INSERT INTO quest_dis_message (
    did, rkey, topic_did, topic_rkey, parent_message_rkey, content, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetMessage :one
SELECT * FROM quest_dis_message
WHERE did = ? AND rkey = ?;

-- name: GetMessagesByTopic :many
SELECT * FROM quest_dis_message
WHERE topic_did = ? AND topic_rkey = ?
ORDER BY created_at ASC;

-- name: GetRepliesByMessage :many
SELECT * FROM quest_dis_message
WHERE topic_did = ? AND topic_rkey = ? AND parent_message_rkey = ?
ORDER BY created_at ASC;

-- name: DeleteMessage :exec
DELETE FROM quest_dis_message
WHERE did = ? AND rkey = ?;

-- Participation queries
-- name: CreateParticipation :one
INSERT INTO quest_dis_participation (
    did, topic_did, topic_rkey, status, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetParticipation :one
SELECT * FROM quest_dis_participation
WHERE did = ? AND topic_did = ? AND topic_rkey = ?;

-- name: GetParticipationsByTopic :many
SELECT * FROM quest_dis_participation
WHERE topic_did = ? AND topic_rkey = ?;

-- name: GetParticipationsByUser :many
SELECT * FROM quest_dis_participation
WHERE did = ?
ORDER BY created_at DESC;

-- name: UpdateParticipationStatus :exec
UPDATE quest_dis_participation
SET status = ?, updated_at = ?
WHERE did = ? AND topic_did = ? AND topic_rkey = ?;

-- name: DeleteParticipation :exec
DELETE FROM quest_dis_participation
WHERE did = ? AND topic_did = ? AND topic_rkey = ?;