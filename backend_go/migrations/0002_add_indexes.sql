-- Add indexes for frequently queried fields

-- Words table indexes
CREATE INDEX IF NOT EXISTS idx_words_tamil ON words(tamil);
CREATE INDEX IF NOT EXISTS idx_words_english ON words(english);

-- Study sessions indexes
CREATE INDEX IF NOT EXISTS idx_study_sessions_created_at ON study_sessions(created_at);

-- Word review items indexes
CREATE INDEX IF NOT EXISTS idx_word_review_items_created_at ON word_review_items(created_at);

-- Words groups indexes
CREATE INDEX IF NOT EXISTS idx_words_groups_word_id ON words_groups(word_id);
CREATE INDEX IF NOT EXISTS idx_words_groups_group_id ON words_groups(group_id); 