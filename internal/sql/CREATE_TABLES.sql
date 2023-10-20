CREATE TABLE products
(
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    stock SMALLINT NOT NULL,
    price REAL NOT NULL,
    discount SMALLINT NOT NULL,
    images VARCHAR[],
    description VARCHAR,
    characteristics bytea,

    subject_id INTEGER REFERENCES subjects (id) ON DELETE CASCADE ON UPDATE CASCADE,
    brand_id INTEGER REFERENCES brands (id) ON DELETE CASCADE ON UPDATE CASCADE
)

CREATE TABLE subjects
(
    id   SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    image VARCHAR
)

CREATE TABLE brands
(
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL
)

CREATE TABLE subjects_brands
(
    id SERIAL PRIMARY KEY,
    subject_id INTEGER REFERENCES subjects (id) ON DELETE CASCADE ON UPDATE CASCADE,
    brand_id INTEGER REFERENCES brands (id) ON DELETE CASCADE ON UPDATE CASCADE
)