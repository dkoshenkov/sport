import {
  benchAssistanceOptions,
  deadliftAssistanceOptions,
  findOption,
  gppAbsOptions,
  gppBicepsOptions,
  gppHorizontalPullOptions,
  gppOverheadPressOptions,
  gppTricepsOptions,
  gppVerticalPullOptions,
  squatAssistanceOptions,
  variants,
  weeks,
} from '../data/programOptions'
import type { LiftId, ProgramSelection, TrainingDay, TrainingRow, VariantId, WeekId } from './types'

const weekOrder: WeekId[] = weeks.map((week) => week.id)

const lightLiftPercent: Record<Exclude<WeekId, 'week_8'>, number> = {
  week_1: 0.62,
  week_2: 0.57,
  week_3: 0.66,
  week_4: 0.64,
  week_5: 0.59,
  week_6: 0.66,
  week_7: 0.7,
}

const lightBenchPercent: Record<Exclude<WeekId, 'week_8'>, number> = {
  week_1: 0.6,
  week_2: 0.65,
  week_3: 0.6,
  week_4: 0.65,
  week_5: 0.65,
  week_6: 0.65,
  week_7: 0.65,
}

const lightLiftPrescription: Record<Exclude<WeekId, 'week_8'>, string> = {
  week_1: '5x5',
  week_2: '4x6',
  week_3: '5x4',
  week_4: '5x5',
  week_5: '4x6',
  week_6: '5x4',
  week_7: '3x3',
}

const kgFormatter = new Intl.NumberFormat('ru-RU', {
  maximumFractionDigits: 1,
})

export function calculateTrainingPlan(selection: ProgramSelection): TrainingDay[] {
  const week = selection.week

  return [
    {
      id: 'day_1',
      title: 'День 1',
      focus: 'Тяжелая становая',
      rows: compactRows([
        heavyRow('day_1_deadlift', 'deadlift', 'Становая тяга', selection.oneRepMax.deadlift, selection),
        lightRow('day_1_bench', 'bench_press', 'Жим лежа', selection.oneRepMax.bench, week, true),
        assistanceRow(
          'day_1_deadlift_assistance',
          findOption(deadliftAssistanceOptions, selection.deadliftAssistance),
          week,
          selection.deadliftAssistance === 'paused_deadlift',
        ),
        gppPullRow('day_1_horizontal_pull', findOption(gppHorizontalPullOptions, selection.gppHorizontalPull), week),
        gppArmRow('day_1_biceps', findOption(gppBicepsOptions, selection.gppBiceps), week),
      ]),
    },
    {
      id: 'day_2',
      title: 'День 2',
      focus: 'Тяжелый жим',
      rows: compactRows([
        heavyRow('day_2_bench', 'bench', 'Жим лежа', selection.oneRepMax.bench, selection),
        lightRow('day_2_squat', 'squat', 'Приседания', selection.oneRepMax.squat, week, false),
        assistanceRow('day_2_bench_assistance', findOption(benchAssistanceOptions, selection.benchAssistance), week),
        gppSimpleRow('day_2_abs', findOption(gppAbsOptions, selection.gppAbs), week, '3x6-10', 'RPE: 7'),
        gppArmRow('day_2_triceps', findOption(gppTricepsOptions, selection.gppTriceps), week),
      ]),
    },
    {
      id: 'day_3',
      title: 'День 3',
      focus: 'Тяжелый присед',
      rows: compactRows([
        heavyRow('day_3_squat', 'squat', 'Приседания', selection.oneRepMax.squat, selection),
        lightRow('day_3_deadlift', 'deadlift', 'Становая тяга', selection.oneRepMax.deadlift, week, false),
        assistanceRow('day_3_squat_assistance', findOption(squatAssistanceOptions, selection.squatAssistance), week),
        gppPullRow('day_3_vertical_pull', findOption(gppVerticalPullOptions, selection.gppVerticalPull), week),
        gppSimpleRow('day_3_overhead_press', findOption(gppOverheadPressOptions, selection.gppOverheadPress), week, '3x6-10', week === 'week_7' ? 'RPE: 5' : 'RPE: 6'),
      ]),
    },
  ]
}

export function roundToNearest2_5(value: number): number {
  return Math.round(value / 2.5) * 2.5
}

export function baseWeight(oneRepMax: number, variant: VariantId): number {
  return oneRepMax * (variant === 'variant_1' ? 0.82 : 0.86)
}

export function progressionStep(selection: ProgramSelection, lift: LiftId): number {
  const percent = selection.progressionStep === 'step_4_percent' ? 0.04 : 0.05
  return roundToNearest2_5(selection.oneRepMax[lift] * percent)
}

export function variantLabel(variant: VariantId) {
  return variants.find((item) => item.id === variant)?.label ?? variant
}

export function weekLabel(week: WeekId) {
  return weeks.find((item) => item.id === week)?.label ?? week
}

function heavyRow(
  id: string,
  lift: LiftId,
  exerciseName: string,
  oneRepMax: number,
  selection: ProgramSelection,
): TrainingRow {
  if (selection.week === 'week_8') {
    return {
      id,
      kind: 'main',
      exerciseKey: lift === 'bench' ? 'bench_press' : lift,
      exerciseName,
      prescription: 'Тест',
      load: '1ПМ',
    }
  }

  const weekIndex = weekOrder.indexOf(selection.week) + 1
  const weeklyOffset = weekIndex - 5
  const weight = roundToNearest2_5(baseWeight(oneRepMax, selection.variant) + progressionStep(selection, lift) * weeklyOffset)

  return {
    id,
    kind: 'main',
    exerciseKey: lift === 'bench' ? 'bench_press' : lift,
    exerciseName,
    prescription: selection.week === 'week_6' ? '3x3' : selection.week === 'week_7' ? '2x2' : '5x5',
    load: formatKg(weight),
    unit: 'т',
  }
}

function lightRow(
  id: string,
  exerciseKey: string,
  exerciseName: string,
  oneRepMax: number,
  week: WeekId,
  isBench: boolean,
): TrainingRow | null {
  if (week === 'week_8') return null

  const percent = isBench ? lightBenchPercent[week] : lightLiftPercent[week]
  const weight = roundToNearest2_5(oneRepMax * percent)

  return {
    id,
    kind: 'light',
    exerciseKey,
    exerciseName,
    prescription: isBench
      ? week === 'week_1' || week === 'week_3'
        ? '4x8'
        : week === 'week_7'
          ? '3x3'
          : '5x5'
      : lightLiftPrescription[week],
    load: formatKg(weight),
    unit: 'л',
  }
}

function assistanceRow(
  id: string,
  option: { label: string; exerciseKey?: string },
  week: WeekId,
  forcePausedDeadliftPattern = false,
): TrainingRow | null {
  if (week === 'week_8' || !option.exerciseKey) return null

  return {
    id,
    kind: 'assistance',
    exerciseKey: option.exerciseKey,
    exerciseName: option.label,
    prescription: forcePausedDeadliftPattern ? '2x4' : week === 'week_6' ? '3x4' : week === 'week_7' ? '2x3' : '2x6',
    load: week === 'week_6' || week === 'week_7' ? 'RPE: 5' : 'RPE: 6-7',
  }
}

function gppPullRow(id: string, option: { id: string; label: string; exerciseKey?: string }, week: WeekId): TrainingRow | null {
  if (week === 'week_8' || option.id === '-' || !option.exerciseKey) return null

  return {
    id,
    kind: 'gpp',
    exerciseKey: option.exerciseKey,
    exerciseName: option.label,
    prescription: week === 'week_7' ? '2x5-6' : '3x6-10',
    load: week === 'week_7' ? 'RPE: 6' : 'RPE: 7',
  }
}

function gppArmRow(id: string, option: { id: string; label: string; exerciseKey?: string }, week: WeekId): TrainingRow | null {
  if (week === 'week_8' || option.id === '-' || !option.exerciseKey) return null

  const setsByWeek: Record<Exclude<WeekId, 'week_8'>, number> = {
    week_1: 2,
    week_2: 3,
    week_3: 3,
    week_4: 4,
    week_5: 2,
    week_6: 3,
    week_7: 3,
  }

  return {
    id,
    kind: 'gpp',
    exerciseKey: option.exerciseKey,
    exerciseName: option.label,
    prescription: `${setsByWeek[week as Exclude<WeekId, 'week_8'>]}x6-12`,
    load: week === 'week_7' ? 'RPE: 6' : 'RPE: 8',
  }
}

function gppSimpleRow(
  id: string,
  option: { id: string; label: string; exerciseKey?: string },
  week: WeekId,
  prescription: string,
  load: string,
): TrainingRow | null {
  if (week === 'week_8' || option.id === '-' || !option.exerciseKey) return null

  return {
    id,
    kind: 'gpp',
    exerciseKey: option.exerciseKey,
    exerciseName: option.label,
    prescription,
    load,
  }
}

function compactRows(rows: Array<TrainingRow | null>): TrainingRow[] {
  return rows.filter((row): row is TrainingRow => row !== null)
}

function formatKg(weight: number): string {
  return `${kgFormatter.format(weight)} кг`
}
