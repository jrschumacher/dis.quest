-- queries.sql - Central SQL query file for dis.quest
-- All SQL queries should be added to this file as documented in CLAUDE.md

-- Topics queries
-- name: CreateTopic :one
INSERT INTO quest_dis_topic (
    did, rkey, subject, initial_message, category, created_at, updated_at, selected_answer
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetTopic :one
SELECT * FROM quest_dis_topic
WHERE did = $1 AND rkey = $2;

-- name: GetTopicsByCategory :many
SELECT * FROM quest_dis_topic
WHERE category = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: ListTopics :many
SELECT * FROM quest_dis_topic
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateTopicSelectedAnswer :exec
UPDATE quest_dis_topic
SET selected_answer = $1, updated_at = $2
WHERE did = $3 AND rkey = $4;

-- name: GetTopicsByDID :many
SELECT * FROM quest_dis_topic
WHERE did = $1
ORDER BY created_at DESC;

-- name: DeleteTopic :exec
DELETE FROM quest_dis_topic
WHERE did = $1 AND rkey = $2;

-- Messages queries
-- name: CreateMessage :one
INSERT INTO quest_dis_message (
    did, rkey, topic_did, topic_rkey, parent_message_rkey, content, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetMessage :one
SELECT * FROM quest_dis_message
WHERE did = $1 AND rkey = $2;

-- name: GetMessagesByTopic :many
SELECT * FROM quest_dis_message
WHERE topic_did = $1 AND topic_rkey = $2
ORDER BY created_at ASC;

-- name: GetRepliesByMessage :many
SELECT * FROM quest_dis_message
WHERE topic_did = $1 AND topic_rkey = $2 AND parent_message_rkey = $3
ORDER BY created_at ASC;

-- name: DeleteMessage :exec
DELETE FROM quest_dis_message
WHERE did = $1 AND rkey = $2;

-- Participation queries
-- name: CreateParticipation :one
INSERT INTO quest_dis_participation (
    did, topic_did, topic_rkey, status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetParticipation :one
SELECT * FROM quest_dis_participation
WHERE did = $1 AND topic_did = $2 AND topic_rkey = $3;

-- name: GetParticipationsByTopic :many
SELECT * FROM quest_dis_participation
WHERE topic_did = $1 AND topic_rkey = $2;

-- name: GetParticipationsByUser :many
SELECT * FROM quest_dis_participation
WHERE did = $1
ORDER BY created_at DESC;

-- name: UpdateParticipationStatus :exec
UPDATE quest_dis_participation
SET status = $1, updated_at = $2
WHERE did = $3 AND topic_did = $4 AND topic_rkey = $5;

-- name: DeleteParticipation :exec
DELETE FROM quest_dis_participation
WHERE did = $1 AND topic_did = $2 AND topic_rkey = $3;