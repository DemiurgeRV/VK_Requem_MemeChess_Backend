ALTER TABLE games
    ADD COLUMN IF NOT EXISTS finished_reason text NULL;

ALTER TABLE games
    ADD COLUMN IF NOT EXISTS paid_out boolean NOT NULL DEFAULT false;

