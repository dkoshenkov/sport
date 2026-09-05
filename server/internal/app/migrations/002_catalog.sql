-- Explicit Unicode case handling, independent of the database's default locale.
CREATE COLLATION sport_unicode (provider = icu, locale = 'und');

CREATE TABLE exercise_catalog (
    id text PRIMARY KEY,
    name text NOT NULL,
    category text NOT NULL DEFAULT '',
    body_part text NOT NULL DEFAULT '',
    equipment text NOT NULL DEFAULT '',
    target text NOT NULL DEFAULT '',
    secondary_muscles text[] NOT NULL DEFAULT '{}',
    instruction_steps jsonb NOT NULL DEFAULT '{}',
    gif_url text NOT NULL DEFAULT '',
    has_media boolean NOT NULL DEFAULT false
);
CREATE INDEX exercise_catalog_name_idx ON exercise_catalog (lower(name), id);
CREATE INDEX exercise_catalog_target_idx ON exercise_catalog (target);
CREATE INDEX exercise_catalog_equipment_idx ON exercise_catalog (equipment);
CREATE TABLE exercise_aliases (
    program_key text PRIMARY KEY,
    program_name text NOT NULL,
    dataset_id text REFERENCES exercise_catalog(id),
    review_status text NOT NULL CHECK (review_status IN ('confirmed', 'missing', 'needs_review')),
    name_hints text[] NOT NULL DEFAULT '{}'
);
CREATE TABLE reference_imports (
    name text PRIMARY KEY,
    imported_at timestamptz NOT NULL DEFAULT now()
);

-- Initial program names and resolution hints. Dataset IDs are resolved after import.
INSERT INTO exercise_aliases(program_key,program_name,review_status,name_hints) VALUES
('deadlift','Становая тяга','confirmed',ARRAY['0032','barbell deadlift']::text[]),
('bench_press','Жим лежа','confirmed',ARRAY['0025','barbell bench press']::text[]),
('squat','Приседания','confirmed',ARRAY['0043','barbell full squat']::text[]),
('pull_up','Подтягивания','confirmed',ARRAY['0652','pull-up']::text[]),
('good_morning','Гуд-морнинг','needs_review',ARRAY['0044','barbell good morning','good morning']::text[]),
('romanian_deadlift','Румынская тяга','needs_review',ARRAY['0085','barbell romanian deadlift','romanian deadlift']::text[]),
('deficit_deadlift','Тяга из ямы','needs_review',ARRAY['deficit deadlift']::text[]),
('sumo_deadlift','Становая тяга сумо','needs_review',ARRAY['sumo deadlift']::text[]),
('paused_deadlift','Становая тяга с паузами','missing',ARRAY[]::text[]),
('close_grip_bench','Жим узким хватом','needs_review',ARRAY['0030','barbell close-grip bench press','close-grip bench press','close grip bench press']::text[]),
('reverse_grip_bench','Жим обратным хватом','needs_review',ARRAY['2187','barbell reverse close-grip bench press','reverse grip bench press']::text[]),
('incline_bench','Жим на наклонной скамье','needs_review',ARRAY['0047','barbell incline bench press','incline bench press']::text[]),
('dumbbell_bench','Жим гантелей лежа','needs_review',ARRAY['0289','dumbbell bench press']::text[]),
('dips','Брусья','needs_review',ARRAY['0251','chest dip','dip']::text[]),
('zercher_squat','Приседания Зерчера','needs_review',ARRAY['1545','barbell full zercher squat','zercher squat']::text[]),
('front_squat','Приседания со штангой на груди','needs_review',ARRAY['0042','barbell front squat','front squat']::text[]),
('high_bar_squat','Приседания с высоким грифом','needs_review',ARRAY['1436','barbell high bar squat','high bar squat']::text[]),
('low_bar_squat','Приседания с низким грифом','needs_review',ARRAY['1435','barbell low bar squat','low bar squat']::text[]),
('bulgarian_split_squat','Болгарские сплит-приседания','needs_review',ARRAY['0410','dumbbell single leg split squat','bulgarian split squat','rear foot elevated split squat']::text[]),
('abs','Упражнение на пресс','missing',ARRAY[]::text[]),
('triceps','Трицепс','missing',ARRAY[]::text[]),
('biceps','Бицепс','confirmed',ARRAY['0294','barbell curl']::text[]),
('barbell_row','Тяга штанги в наклоне','needs_review',ARRAY['0027','barbell bent over row','bent over barbell row']::text[]),
('cable_seated_row','Горизонтальный блок','needs_review',ARRAY['seated cable row','cable seated row']::text[]),
('dumbbell_row','Тяга гантели в наклоне','needs_review',ARRAY['one arm dumbbell row','dumbbell row']::text[]),
('lever_horizontal_row','Рычажная горизонтальная тяга','needs_review',ARRAY['lever seated row','lever row']::text[]),
('lat_pulldown','Вертикальный блок','needs_review',ARRAY['0150','cable bar lateral pulldown','cable pulldown','lat pulldown']::text[]),
('lever_vertical_row','Рычажная вертикальная тяга','needs_review',ARRAY['0579','lever front pulldown','lever pulldown','lever vertical row']::text[]),
('dumbbell_military_press','Армейский жим гантелей','needs_review',ARRAY['0405','dumbbell seated shoulder press','dumbbell shoulder press','dumbbell overhead press']::text[]),
('handstand_push_up','Отжимания в стойке на руках','needs_review',ARRAY['handstand push-up']::text[]),
('kettlebell_military_press','Армейский жим гирь','needs_review',ARRAY['0553','kettlebell two arm military press','kettlebell clean and press','kettlebell press']::text[]),
('one_arm_military_press','Армейский жим одной рукой','needs_review',ARRAY['0361','dumbbell one arm shoulder press','one arm dumbbell press','single arm shoulder press']::text[]),
('barbell_military_press','Армейский жим штанги','needs_review',ARRAY['1456','barbell standing close grip military press','barbell standing military press','barbell shoulder press']::text[]);
