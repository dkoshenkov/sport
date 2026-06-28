import type { ProgramSelection, VariantId, WeekId } from '../domain/types'

export type ExerciseOption = {
  id: string
  label: string
  exerciseKey?: string
}

export const weeks: Array<{ id: WeekId; label: string }> = [
  { id: 'week_1', label: 'Неделя 1' },
  { id: 'week_2', label: 'Неделя 2' },
  { id: 'week_3', label: 'Неделя 3' },
  { id: 'week_4', label: 'Неделя 4' },
  { id: 'week_5', label: 'Неделя 5' },
  { id: 'week_6', label: 'Неделя 6' },
  { id: 'week_7', label: 'Неделя 7' },
  { id: 'week_8', label: 'Неделя 8' },
]

export const variants: Array<{ id: VariantId; label: string; note: string }> = [
  { id: 'variant_1', label: 'Вариант 1', note: '0.82 от 1ПМ' },
  { id: 'variant_2', label: 'Вариант 2', note: '0.86 от 1ПМ' },
]

export const deadliftAssistanceOptions: ExerciseOption[] = [
  { id: 'good_morning', label: 'Гуд-морнинг', exerciseKey: 'good_morning' },
  { id: 'romanian_deadlift', label: 'Румынская тяга', exerciseKey: 'romanian_deadlift' },
  { id: 'deficit_deadlift', label: 'Тяга из ямы', exerciseKey: 'deficit_deadlift' },
  { id: 'classic_deadlift', label: 'Классическая становая тяга', exerciseKey: 'deadlift' },
  { id: 'sumo_deadlift', label: 'Становая тяга сумо', exerciseKey: 'sumo_deadlift' },
  { id: 'paused_deadlift', label: 'Становя тяга с паузами', exerciseKey: 'paused_deadlift' },
]

export const benchAssistanceOptions: ExerciseOption[] = [
  { id: 'close_grip_bench', label: 'Жим узким хватом', exerciseKey: 'close_grip_bench' },
  { id: 'reverse_grip_bench', label: 'Жим обратным хватом', exerciseKey: 'reverse_grip_bench' },
  { id: 'incline_bench', label: 'Жим на наклонной скамье', exerciseKey: 'incline_bench' },
  { id: 'dumbbell_bench', label: 'Жим гантелей лежа', exerciseKey: 'dumbbell_bench' },
  { id: 'dips', label: 'Брусья', exerciseKey: 'dips' },
]

export const squatAssistanceOptions: ExerciseOption[] = [
  { id: 'zercher_squat', label: 'Приседания Зерчера', exerciseKey: 'zercher_squat' },
  { id: 'front_squat', label: 'Приседания со штангой на груди', exerciseKey: 'front_squat' },
  { id: 'high_bar_squat', label: 'Приседания с высоким грифом', exerciseKey: 'high_bar_squat' },
  { id: 'low_bar_squat', label: 'Приседания с низком грифом', exerciseKey: 'low_bar_squat' },
  { id: 'bulgarian_split_squat', label: 'Болгарские сплит-приседания', exerciseKey: 'bulgarian_split_squat' },
]

export const gppHorizontalPullOptions: ExerciseOption[] = [
  { id: '-', label: '-' },
  { id: 'barbell_row', label: 'Тяга штанги в наклоне', exerciseKey: 'barbell_row' },
  { id: 'cable_seated_row', label: 'Горизонтальный блок', exerciseKey: 'cable_seated_row' },
  { id: 'dumbbell_row', label: 'Тяга гантели в наклоне', exerciseKey: 'dumbbell_row' },
  { id: 'lever_horizontal_row', label: 'Рычажная горизонтальная тяга', exerciseKey: 'lever_horizontal_row' },
]

export const gppBicepsOptions: ExerciseOption[] = [
  { id: '-', label: '-' },
  { id: 'biceps', label: 'Бицепс', exerciseKey: 'biceps' },
]

export const gppAbsOptions: ExerciseOption[] = [
  { id: '-', label: '-' },
  { id: 'abs', label: 'Упражнение на пресс', exerciseKey: 'abs' },
]

export const gppTricepsOptions: ExerciseOption[] = [
  { id: '-', label: '-' },
  { id: 'triceps', label: 'Трицепс', exerciseKey: 'triceps' },
]

export const gppVerticalPullOptions: ExerciseOption[] = [
  { id: '-', label: '-' },
  { id: 'pull_up', label: 'Подтягивания', exerciseKey: 'pull_up' },
  { id: 'lat_pulldown', label: 'Вертикальный блок', exerciseKey: 'lat_pulldown' },
  { id: 'lever_vertical_row', label: 'Рычажная вертикальная тяга', exerciseKey: 'lever_vertical_row' },
]

export const gppOverheadPressOptions: ExerciseOption[] = [
  { id: '-', label: '-' },
  { id: 'dumbbell_military_press', label: 'Армейский жим гантелей', exerciseKey: 'dumbbell_military_press' },
  { id: 'handstand_push_up', label: 'Отжимания в стойке на руках', exerciseKey: 'handstand_push_up' },
  { id: 'kettlebell_military_press', label: 'Армейский жим гирь', exerciseKey: 'kettlebell_military_press' },
  { id: 'one_arm_military_press', label: 'Армейский жим одной рукой', exerciseKey: 'one_arm_military_press' },
  { id: 'barbell_military_press', label: 'Армейский жим штанги', exerciseKey: 'barbell_military_press' },
]

export const defaultSelection: ProgramSelection = {
  oneRepMax: {
    deadlift: 225,
    bench: 125,
    squat: 170,
  },
  week: 'week_3',
  variant: 'variant_1',
  progressionStep: 'step_4_percent',
  deadliftAssistance: 'good_morning',
  benchAssistance: 'close_grip_bench',
  squatAssistance: 'front_squat',
  gppHorizontalPull: 'barbell_row',
  gppBiceps: 'biceps',
  gppAbs: 'abs',
  gppTriceps: 'triceps',
  gppVerticalPull: 'pull_up',
  gppOverheadPress: 'kettlebell_military_press',
}

export function findOption(options: ExerciseOption[], id: string): ExerciseOption {
  return options.find((option) => option.id === id) ?? options[0]
}
