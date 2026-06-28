export type LiftId = 'deadlift' | 'bench' | 'squat'

export type WeekId =
  | 'week_1'
  | 'week_2'
  | 'week_3'
  | 'week_4'
  | 'week_5'
  | 'week_6'
  | 'week_7'
  | 'week_8'

export type VariantId = 'variant_1' | 'variant_2'

export type ProgressionStepId = 'step_4_percent' | 'step_5_percent'

export type ProgramSelection = {
  oneRepMax: Record<LiftId, number>
  week: WeekId
  variant: VariantId
  progressionStep: ProgressionStepId
  deadliftAssistance: string
  benchAssistance: string
  squatAssistance: string
  gppHorizontalPull: string
  gppBiceps: string
  gppAbs: string
  gppTriceps: string
  gppVerticalPull: string
  gppOverheadPress: string
}

export type TrainingRowKind = 'main' | 'light' | 'assistance' | 'gpp'

export type TrainingRow = {
  id: string
  kind: TrainingRowKind
  exerciseKey: string
  exerciseName: string
  prescription: string
  load: string
  unit?: string
}

export type TrainingDay = {
  id: 'day_1' | 'day_2' | 'day_3'
  title: string
  focus: string
  rows: TrainingRow[]
}
