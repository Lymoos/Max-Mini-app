CREATE TABLE user_benefits (
    user_id      TEXT NOT NULL,
    benefit_id   TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'not_started'
        CHECK (status IN ('not_started', 'collecting', 'submitted', 'review', 'approved', 'rejected')),
    submitted_at DATE,
    docs         INT[] NOT NULL DEFAULT '{}',
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, benefit_id)
);

CREATE TABLE user_guides (
    user_id  TEXT NOT NULL,
    guide_id TEXT NOT NULL,
    read_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, guide_id)
);
