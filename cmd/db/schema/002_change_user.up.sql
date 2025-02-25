ALTER TABLE users ADD COLUMN profile_picture TEXT NOT NULL;
ALTER TABLE users ADD COLUMN confirmed_email boolean DEFAULT 'f' NOT NULL;
ALTER TABLE users ADD COLUMN username varchar(32) UNIQUE NOT NULL;

CREATE TABLE tokens(
    id UUID NOT NULL PRIMARY KEY,
    user_id UUID REFERENCES users(id) NOT NULL,
    token varchar(64) NOT NULL,
    exp_date Timestamp NOT NULL
);