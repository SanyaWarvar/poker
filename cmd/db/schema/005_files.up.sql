CREATE TABLE files(
    file_data TEXT NOT NULL,
    file_path varchar(255) UNIQUE NOT NULL
);

ALTER TABLE users 
ALTER COLUMN profile_picture 
TYPE varchar(255);

ALTER TABLE users
DROP COLUMN pic_ext;