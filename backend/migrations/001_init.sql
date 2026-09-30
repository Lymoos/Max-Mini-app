CREATE TABLE profiles (
    user_id       TEXT PRIMARY KEY,
    name          TEXT NOT NULL DEFAULT '',
    birth_date    DATE,
    address       TEXT NOT NULL DEFAULT '',
    lat           DOUBLE PRECISION,
    lon           DOUBLE PRECISION,
    contact_name  TEXT NOT NULL DEFAULT '',
    contact_phone TEXT NOT NULL DEFAULT '',
    health        TEXT NOT NULL DEFAULT '',
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE tasks (
    id      SERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    date    DATE NOT NULL,
    time    TEXT NOT NULL,
    title   TEXT NOT NULL,
    kind    TEXT NOT NULL,
    done    BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX tasks_user_date_idx ON tasks (user_id, date);

-- аптеки и магазины, выгруженные из OpenStreetMap
CREATE TABLE places (
    id                BIGSERIAL PRIMARY KEY,
    osm_id            TEXT NOT NULL UNIQUE,
    kind              TEXT NOT NULL CHECK (kind IN ('pharmacy', 'shop')),
    name              TEXT NOT NULL,
    address           TEXT NOT NULL DEFAULT '',
    city              TEXT NOT NULL,
    lat               DOUBLE PRECISION NOT NULL,
    lon               DOUBLE PRECISION NOT NULL,
    rating            REAL NOT NULL,
    reviews           INT NOT NULL,
    birthday_discount INT NOT NULL DEFAULT 0,
    source            TEXT NOT NULL DEFAULT 'demo'
);

CREATE INDEX places_kind_lat_lon_idx ON places (kind, lat, lon);

CREATE TABLE medicines (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    form       TEXT NOT NULL,
    base_price INT NOT NULL
);

CREATE TABLE medicine_prices (
    place_id    BIGINT NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    medicine_id TEXT NOT NULL REFERENCES medicines (id) ON DELETE CASCADE,
    price       INT NOT NULL,
    source      TEXT NOT NULL DEFAULT 'demo',
    PRIMARY KEY (place_id, medicine_id)
);

CREATE INDEX medicine_prices_medicine_idx ON medicine_prices (medicine_id);

CREATE TABLE products (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    unit       TEXT NOT NULL,
    base_price INT NOT NULL
);

CREATE TABLE product_prices (
    place_id    BIGINT NOT NULL REFERENCES places (id) ON DELETE CASCADE,
    product_id  TEXT NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    price       INT NOT NULL,
    promo_price INT,
    promo_until DATE,
    source      TEXT NOT NULL DEFAULT 'demo',
    PRIMARY KEY (place_id, product_id)
);

CREATE INDEX product_prices_product_idx ON product_prices (product_id);

CREATE TABLE ai_cache (
    key        TEXT PRIMARY KEY,
    answer     TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
