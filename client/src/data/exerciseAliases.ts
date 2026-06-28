import { exerciseDetailsByDatasetId } from './exerciseDetails'

export type AliasStatus = 'confirmed' | 'missing' | 'needs_review'

export type ExerciseAlias = {
  programExerciseKey: string
  programNameRu: string
  datasetExerciseId: string | null
  datasetName: string | null
  status: AliasStatus
  notes?: string
}

export const exerciseAliases: Record<string, ExerciseAlias> = {
  deadlift: alias('deadlift', 'Становая тяга', '0032'),
  bench_press: alias('bench_press', 'Жим лежа', '0025'),
  squat: alias('squat', 'Приседания', '1461', 'Dataset entry is generic back-view barbell full squat.'),
  good_morning: alias('good_morning', 'Гуд-морнинг', '0044'),
  romanian_deadlift: alias('romanian_deadlift', 'Румынская тяга', '0085'),
  deficit_deadlift: review('deficit_deadlift', 'Тяга из ямы', 'No exact deficit deadlift entry found in dataset search.'),
  sumo_deadlift: alias('sumo_deadlift', 'Становая тяга сумо', '0117'),
  paused_deadlift: review('paused_deadlift', 'Становя тяга с паузами', 'No exact paused deadlift entry found; XLSX also contains this typo.'),
  close_grip_bench: alias('close_grip_bench', 'Жим узким хватом', '0030'),
  reverse_grip_bench: alias('reverse_grip_bench', 'Жим обратным хватом', '1258', 'Closest confirmed dataset entry uses wide reverse grip.'),
  incline_bench: alias('incline_bench', 'Жим на наклонной скамье', '0047'),
  dumbbell_bench: alias('dumbbell_bench', 'Жим гантелей лежа', '0289'),
  dips: alias('dips', 'Брусья', '0251', 'Mapped to chest dip; XLSX does not specify chest/triceps bias.'),
  zercher_squat: alias('zercher_squat', 'Приседания Зерчера', '0127'),
  front_squat: alias('front_squat', 'Приседания со штангой на груди', '0042'),
  high_bar_squat: review('high_bar_squat', 'Приседания с высоким грифом', 'Dataset search did not confirm high-bar vs low-bar distinction.'),
  low_bar_squat: review('low_bar_squat', 'Приседания с низком грифом', 'Dataset search did not confirm high-bar vs low-bar distinction.'),
  bulgarian_split_squat: alias('bulgarian_split_squat', 'Болгарские сплит-приседания', '0099', 'Closest confirmed dataset entry is barbell single leg split squat.'),
  barbell_row: alias('barbell_row', 'Тяга штанги в наклоне', '0027'),
  cable_seated_row: alias('cable_seated_row', 'Горизонтальный блок', '0861'),
  dumbbell_row: alias('dumbbell_row', 'Тяга гантели в наклоне', '0293'),
  lever_horizontal_row: alias('lever_horizontal_row', 'Рычажная горизонтальная тяга', '0571'),
  biceps: alias('biceps', 'Бицепс', '0285', 'Mapped to dumbbell alternate biceps curl; XLSX uses a generic local/isolation group.'),
  abs: missing('abs', 'Упражнение на пресс', 'XLSX uses a generic abs placeholder, not a concrete exercise.'),
  triceps: alias('triceps', 'Трицепс', '0060', 'Mapped to skull crusher; XLSX uses a generic local/isolation group.'),
  pull_up: alias('pull_up', 'Подтягивания', '0140', 'Dataset match is biceps pull-up; generic pull-up needs review if grip matters.'),
  lat_pulldown: alias('lat_pulldown', 'Вертикальный блок', '2330'),
  lever_vertical_row: review('lever_vertical_row', 'Рычажная вертикальная тяга', 'No exact lever vertical row entry confirmed.'),
  dumbbell_military_press: alias('dumbbell_military_press', 'Армейский жим гантелей', '0405'),
  handstand_push_up: alias('handstand_push_up', 'Отжимания в стойке на руках', '0471'),
  kettlebell_military_press: alias('kettlebell_military_press', 'Армейский жим гирь', '0553'),
  one_arm_military_press: alias('one_arm_military_press', 'Армейский жим одной рукой', '0539', 'Mapped to kettlebell one-arm military press; implement equipment-specific choice later.'),
  barbell_military_press: alias('barbell_military_press', 'Армейский жим штанги', '1457'),
}

function alias(
  programExerciseKey: string,
  programNameRu: string,
  datasetExerciseId: string,
  notes?: string,
): ExerciseAlias {
  return {
    programExerciseKey,
    programNameRu,
    datasetExerciseId,
    datasetName: exerciseDetailsByDatasetId[datasetExerciseId]?.datasetName ?? null,
    status: 'confirmed',
    notes,
  }
}

function missing(programExerciseKey: string, programNameRu: string, notes: string): ExerciseAlias {
  return {
    programExerciseKey,
    programNameRu,
    datasetExerciseId: null,
    datasetName: null,
    status: 'missing',
    notes,
  }
}

function review(programExerciseKey: string, programNameRu: string, notes: string): ExerciseAlias {
  return {
    programExerciseKey,
    programNameRu,
    datasetExerciseId: null,
    datasetName: null,
    status: 'needs_review',
    notes,
  }
}
