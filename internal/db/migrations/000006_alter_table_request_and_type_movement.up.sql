update request set movement_type_id = 1 where movement_type_id is null;

ALTER TABLE request ALTER COLUMN movement_type_id SET NOT NULL;