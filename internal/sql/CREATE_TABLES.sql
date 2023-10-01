CREATE TABLE products
(
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    stock SMALLINT NOT NULL,
    price REAL NOT NULL,
    discount SMALLINT NOT NULL,
    images VARCHAR[],
    description VARCHAR,
    characteristics bytea
)