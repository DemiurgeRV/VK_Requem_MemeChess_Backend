INSERT INTO users (id, username, email)
VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'user1', 'user1@test.local'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'user2', 'user2@test.local')
    ON CONFLICT (id) DO NOTHING;