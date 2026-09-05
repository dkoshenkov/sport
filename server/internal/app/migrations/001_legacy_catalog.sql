-- Must run before 002_catalog.sql, including on installations without migration history.
-- Renaming columns preserves primary keys and incoming/outgoing foreign keys.
DO $migration$
DECLARE
    mapping record;
    relation regclass;
BEGIN
    FOR mapping IN SELECT * FROM (VALUES
        ('exercise_catalog', 'dataset_exercise_id', 'id'),
        ('exercise_aliases', 'program_exercise_key', 'program_key'),
        ('exercise_aliases', 'program_name_ru', 'program_name'),
        ('exercise_aliases', 'dataset_exercise_id', 'dataset_id')
    ) AS names(table_name, old_name, new_name)
    LOOP
        relation := to_regclass(format('%I.%I', current_schema(), mapping.table_name));
        IF relation IS NOT NULL AND EXISTS (
            SELECT 1 FROM pg_attribute WHERE attrelid = relation
                AND attname = mapping.old_name AND NOT attisdropped
        ) THEN
            IF EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = relation
                AND attname = mapping.new_name AND NOT attisdropped) THEN
                RAISE EXCEPTION 'Cannot migrate %.%: column % also exists',
                    mapping.table_name, mapping.old_name, mapping.new_name;
            END IF;
            EXECUTE format('ALTER TABLE %s RENAME COLUMN %I TO %I',
                relation, mapping.old_name, mapping.new_name);
        END IF;
    END LOOP;

    relation := to_regclass(format('%I.exercise_catalog', current_schema()));
    IF relation IS NOT NULL THEN
        ALTER TABLE exercise_catalog ADD COLUMN IF NOT EXISTS gif_url text NOT NULL DEFAULT '';
        ALTER TABLE exercise_catalog ADD COLUMN IF NOT EXISTS has_media boolean NOT NULL DEFAULT false;
        IF EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = relation
            AND attname = 'gif_path' AND NOT attisdropped) THEN
            UPDATE exercise_catalog SET gif_url = gif_path
            WHERE gif_url = '' AND gif_path IS NOT NULL;
        END IF;
    END IF;
END
$migration$;
