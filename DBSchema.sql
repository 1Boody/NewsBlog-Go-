CREATE TABLE users (
    name text,
    email text,
    id SERIAL,
    password text,
    PRIMARY KEY(id)
)

CREATE TABLE blogs(
    blog_id SERIAL,
    author_id int,
    title text,
    body text,
    url SERIAL,
    PRIMARY KEY(blog_id),
    FOREIGN KEY(author_id) REFERENCES users(id)
)