DROP TABLE files;

ALTER TABLE users 
DROP COLUMN profile_picture;

ALTER TABLE users 
ADD COLUMN pic_ext varchar(6) 
DEFAULT '.jpg' NOT NULL;