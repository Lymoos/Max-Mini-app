ALTER TABLE places DROP CONSTRAINT places_kind_check;
ALTER TABLE places ADD CONSTRAINT places_kind_check CHECK (kind IN ('pharmacy', 'shop', 'clinic', 'social'));
ALTER TABLE places ADD COLUMN phone TEXT NOT NULL DEFAULT '';
ALTER TABLE places ADD COLUMN is_state BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE profiles ADD COLUMN reg_address TEXT NOT NULL DEFAULT '';
ALTER TABLE profiles ADD COLUMN reg_lat DOUBLE PRECISION;
ALTER TABLE profiles ADD COLUMN reg_lon DOUBLE PRECISION;
ALTER TABLE profiles ADD COLUMN clinic_id BIGINT REFERENCES places (id) ON DELETE SET NULL;

-- списков врачей в открытом доступе нет, поэтому они сгенерированы и помечены source = 'demo'
CREATE TABLE doctors (
    id         BIGSERIAL PRIMARY KEY,
    place_id   BIGINT NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    specialty  TEXT NOT NULL,
    experience INT NOT NULL,
    category   TEXT NOT NULL DEFAULT '',
    rating     REAL NOT NULL,
    reviews    INT NOT NULL,
    source     TEXT NOT NULL DEFAULT 'demo'
);

CREATE INDEX doctors_place_idx ON doctors (place_id);
