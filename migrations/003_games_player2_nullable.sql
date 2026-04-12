-- Invite links: second player is assigned only after they join.
ALTER TABLE games
    ALTER COLUMN player2_id DROP NOT NULL;
