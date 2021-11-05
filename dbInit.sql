CREATE TABLE IF NOT EXISTS USERS (
    id SERIAL NOT NULL PRIMARY KEY,
    username VARCHAR(40) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    profile_pic VARCHAR(255)
);
CREATE TABLE IF NOT EXISTS CITY (
    id SERIAL NOT NULL PRIMARY KEY,
    name VARCHAR(40) NOT NULL,
    country VARCHAR(40) NOT NULL
);
CREATE TABLE IF NOT EXISTS ITINERARY (
    id SERIAL NOT NULL PRIMARY KEY,
    title VARCHAR(40) NOT NULL,
    creator INTEGER REFERENCES USERS(id) NOT NULL,
    time VARCHAR(40) NOT NULL,
    price VARCHAR(40) NOT NULL,
    activities VARCHAR(40) [50] NOT NULL,
    hashtags VARCHAR(40) [3],
    city_id INTEGER NOT NULL REFERENCES CITY(id)
);
CREATE TABLE IF NOT EXISTS ITINERARY_COMMENT (
    id SERIAL NOT NULL PRIMARY KEY,
    author_id INTEGER REFERENCES USERS(id) NOT NULL,
    comment VARCHAR(255) NOT NULL
);
-- Junction table
CREATE TABLE IF NOT EXISTS ITINERARY_COMMENTS (
    id SERIAL NOT NULL PRIMARY KEY,
    itinerary_id INTEGER NOT NULL REFERENCES ITINERARY(id),
    comment_id INTEGER NOT NULL REFERENCES ITINERARY_COMMENT(id)
);
create table if not exists SESSIONS (
    id serial not null primary key,
    user_id INT not null references users(id),
    session_id TEXT not null unique,
    expiration TIMESTAMP with time zone not null
);