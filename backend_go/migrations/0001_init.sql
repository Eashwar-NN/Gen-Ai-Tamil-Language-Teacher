-- Enable foreign key support
PRAGMA foreign_keys = ON;

-- Enable UTF-8 encoding
PRAGMA encoding = 'UTF-8';

-- Create words table
CREATE TABLE IF NOT EXISTS words (
    id INTEGER PRIMARY KEY,
    tamil TEXT NOT NULL COLLATE NOCASE,
    romaji TEXT NOT NULL COLLATE NOCASE,
    english TEXT NOT NULL COLLATE NOCASE,
    parts TEXT
) STRICT;

-- Create groups table
CREATE TABLE IF NOT EXISTS groups (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL COLLATE NOCASE
) STRICT;

-- Create words_groups join table
CREATE TABLE IF NOT EXISTS words_groups (
    id INTEGER PRIMARY KEY,
    word_id INTEGER,
    group_id INTEGER
) STRICT;

-- Create study_activities table
CREATE TABLE IF NOT EXISTS study_activities (
    id INTEGER PRIMARY KEY,
    group_id INTEGER,
    name TEXT NOT NULL DEFAULT '' COLLATE NOCASE,
    thumbnail_url TEXT,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) STRICT;

-- Create study_sessions table
CREATE TABLE IF NOT EXISTS study_sessions (
    id INTEGER PRIMARY KEY,
    group_id INTEGER,
    study_activities_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) STRICT;

-- Create word_review_items table
CREATE TABLE IF NOT EXISTS word_review_items (
    id INTEGER PRIMARY KEY,
    word_id INTEGER,
    study_session_id INTEGER,
    correct BOOLEAN,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) STRICT;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_words_groups_word_id ON words_groups(word_id);
CREATE INDEX IF NOT EXISTS idx_words_groups_group_id ON words_groups(group_id);
CREATE INDEX IF NOT EXISTS idx_study_activities_group_id ON study_activities(group_id);
CREATE INDEX IF NOT EXISTS idx_study_sessions_group_id ON study_sessions(group_id);
CREATE INDEX IF NOT EXISTS idx_study_sessions_activity_id ON study_sessions(study_activities_id);
CREATE INDEX IF NOT EXISTS idx_word_review_items_word_id ON word_review_items(word_id);
CREATE INDEX IF NOT EXISTS idx_word_review_items_session_id ON word_review_items(study_session_id); 