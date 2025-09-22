-- Verifica si la columna existe antes de eliminarla
DO $$
BEGIN
    IF EXISTS (
        SELECT FROM information_schema.columns
        WHERE table_name='documents' AND column_name='bedrock_id'
    ) THEN
        ALTER TABLE documents DROP COLUMN bedrock_id;
    END IF;
END $$;