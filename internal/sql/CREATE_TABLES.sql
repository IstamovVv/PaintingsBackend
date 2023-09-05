CREATE TABLE products
(
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    images VARCHAR[],
    stock SMALLINT,
    discount SMALLINT,
    description VARCHAR,
    characteristics bytea
)