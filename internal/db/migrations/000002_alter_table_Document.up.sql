ALTER TABLE documents ADD COLUMN IF NOT EXISTS textract_id varchar(255) default null;

ALTER TABLE documents ADD COLUMN IF NOT EXISTS bedrock_id varchar(255) default null;