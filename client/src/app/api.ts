export type LiftId = 'deadlift' | 'bench' | 'squat'
export type ProgramWeek = 'week_1' | 'week_2' | 'week_3' | 'week_4' | 'week_5' | 'week_6' | 'week_7' | 'week_8'
export type ProgramVariant = 'variant_1' | 'variant_2'
export type ProgressionStep = 'step_4_percent' | 'step_5_percent'
export type TrainingRowKind = 'main' | 'light' | 'assistance' | 'gpp'

export type User = {
  id: string
  nickname: string
  createdAt: string
}

export type AthleteProfile = {
  oneRepMaxKg: Record<LiftId, number | null>
  preferredVariant: ProgramVariant
  preferredProgressionStep: ProgressionStep
  notes?: string | null
  createdAt: string
  updatedAt: string
}

export type SelectOption = {
  id: string
  label: string
}

export type ProgramOptions = {
  weeks: SelectOption[]
  variants: SelectOption[]
  progressionSteps: SelectOption[]
  assistance: {
    deadlift: SelectOption[]
    bench: SelectOption[]
    squat: SelectOption[]
  }
  gpp: {
    abs: SelectOption[]
    triceps: SelectOption[]
    horizontalPull: SelectOption[]
    biceps: SelectOption[]
    verticalPull: SelectOption[]
    overheadPress: SelectOption[]
  }
}

export type CycleSettings = {
  oneRepMaxKg: Record<LiftId, number>
  variant: ProgramVariant
  progressionStep: ProgressionStep
  assistance: {
    deadlift: string
    bench: string
    squat: string
  }
  gpp: {
    abs: string | null
    triceps: string | null
    horizontalPull: string | null
    biceps: string | null
    verticalPull: string | null
    overheadPress: string | null
  }
}

export type ProgramCycle = {
  id: string
  title: string
  status: 'active' | 'completed' | 'archived'
  currentWeek: ProgramWeek
  settings: CycleSettings
  progressSummary: {
    done: number
    partial: number
    skipped: number
    planned: number
  }
  createdAt: string
  updatedAt: string
}

export type CyclesResponse = {
  cycles: ProgramCycle[]
  currentCycleId?: string | null
}

export type TrainingPlan = {
  selection: {
    settings: CycleSettings
    week: ProgramWeek
  }
  days: TrainingDay[]
  formulaVersion: string
  warnings: string[]
}

export type TrainingDay = {
  id: 'day_1' | 'day_2' | 'day_3'
  label: string
  focus: string
  rows: TrainingRow[]
}

export type TrainingRow = {
  rowId: string
  exerciseKey: string
  exerciseName: string
  kind: TrainingRowKind
  prescription: {
    setsRepsText: string
    weightKg?: number | null
    weightText?: string | null
    rpeText?: string | null
    unit?: string | null
  }
}

export type ExerciseDetails = {
  exerciseKey: string
  name: string
  datasetExerciseId?: string | null
  datasetName?: string | null
  aliasStatus: 'confirmed' | 'missing' | 'needs_review'
  equipment?: string | null
  targetMuscles?: string[]
  secondaryMuscles?: string[]
  instructions?: string[]
  media: {
    status: 'available' | 'missing'
    gifUrl?: string | null
    width?: number | null
    height?: number | null
  }
}

export type AuthSession = {
  authenticated: boolean
  user?: User
}

export type ExerciseCatalogItem = {
  datasetExerciseId: string
  name: string
  nameRu?: string | null
  category?: string | null
  bodyPart?: string | null
  equipment?: string | null
  targetMuscles: string[]
  secondaryMuscles: string[]
  instructions: string[]
  media: {
    status: 'available' | 'missing'
    gifUrl?: string | null
    width?: number | null
    height?: number | null
  }
}

export type ExerciseCatalogResponse = {
  items: ExerciseCatalogItem[]
  total: number
  limit: number
  offset: number
}

export class ApiError extends Error {
  status: number
  code?: string

  constructor(status: number, message: string, code?: string) {
    super(message)
    this.status = status
    this.code = code
  }
}

const apiBase = import.meta.env.VITE_API_BASE_URL ?? ''

export const defaultCycleSettings: CycleSettings = {
  oneRepMaxKg: {
    deadlift: 225,
    bench: 125,
    squat: 170,
  },
  variant: 'variant_1',
  progressionStep: 'step_4_percent',
  assistance: {
    deadlift: 'good_morning',
    bench: 'close_grip_bench',
    squat: 'front_squat',
  },
  gpp: {
    abs: 'abs',
    triceps: 'triceps',
    horizontalPull: 'barbell_row',
    biceps: 'biceps',
    verticalPull: 'pull_up',
    overheadPress: 'kettlebell_military_press',
  },
}

export function settingsFromProfile(profile: AthleteProfile): CycleSettings {
  return {
    ...defaultCycleSettings,
    oneRepMaxKg: {
      deadlift: profile.oneRepMaxKg.deadlift ?? defaultCycleSettings.oneRepMaxKg.deadlift,
      bench: profile.oneRepMaxKg.bench ?? defaultCycleSettings.oneRepMaxKg.bench,
      squat: profile.oneRepMaxKg.squat ?? defaultCycleSettings.oneRepMaxKg.squat,
    },
    variant: profile.preferredVariant,
    progressionStep: profile.preferredProgressionStep,
  }
}

export async function bootstrap() {
  return request('/v1/bootstrap')
}

export async function getMe(): Promise<User> {
  const data = await request<{ user: User }>('/v1/me')
  return data.user
}

export async function getAuthSession(): Promise<AuthSession> {
  return request<AuthSession>('/v1/auth/session')
}

export async function login(nickname: string, password: string): Promise<User> {
  const data = await request<{ user: User }>('/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ nickname, password }),
  })
  return data.user
}

export async function register(nickname: string, password: string): Promise<User> {
  const data = await request<{ user: User }>('/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify({ nickname, password }),
  })
  return data.user
}

export async function logout() {
  await request('/v1/auth/logout', { method: 'POST' })
}

export async function getProfile(): Promise<AthleteProfile> {
  const data = await request<{ profile: AthleteProfile }>('/v1/me/profile')
  return data.profile
}

export async function getProgramOptions(): Promise<ProgramOptions> {
  return request<ProgramOptions>('/v1/program/options')
}

export async function listCycles(): Promise<CyclesResponse> {
  return request<CyclesResponse>('/v1/cycles')
}

export async function getCurrentCycle(): Promise<ProgramCycle | null> {
  try {
    const data = await request<{ cycle: ProgramCycle }>('/v1/cycles/current')
    return data.cycle
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) return null
    throw error
  }
}

export async function createCycle(title: string, currentWeek: ProgramWeek, settings: CycleSettings): Promise<ProgramCycle> {
  const data = await request<{ cycle: ProgramCycle }>('/v1/cycles', {
    method: 'POST',
    body: JSON.stringify({ title, currentWeek, settings }),
  })
  return data.cycle
}

export async function saveCurrentCycle(title: string, currentWeek: ProgramWeek, settings: CycleSettings): Promise<ProgramCycle> {
  const data = await request<{ cycle: ProgramCycle }>('/v1/cycles/current', {
    method: 'PUT',
    body: JSON.stringify({ title, currentWeek, settings }),
  })
  return data.cycle
}

export async function saveCycle(cycleId: string, title: string, currentWeek: ProgramWeek, settings: CycleSettings): Promise<ProgramCycle> {
  const data = await request<{ cycle: ProgramCycle }>(`/v1/cycles/${encodeURIComponent(cycleId)}`, {
    method: 'PUT',
    body: JSON.stringify({ title, currentWeek, settings }),
  })
  return data.cycle
}

export async function activateCycle(cycleId: string): Promise<ProgramCycle> {
  const data = await request<{ cycle: ProgramCycle }>(`/v1/cycles/${encodeURIComponent(cycleId)}/activate`, {
    method: 'POST',
  })
  return data.cycle
}

export async function getCurrentCyclePlan(): Promise<TrainingPlan> {
  return request<TrainingPlan>('/v1/cycles/current/plan')
}

export async function getExerciseDetails(exerciseKey: string): Promise<ExerciseDetails> {
  const data = await request<{ exercise: ExerciseDetails }>(`/v1/exercises/${encodeURIComponent(exerciseKey)}`)
  return data.exercise
}

export async function listExercises(params: { query?: string; hasGif?: boolean; limit?: number; offset?: number } = {}): Promise<ExerciseCatalogResponse> {
  const search = new URLSearchParams()
  if (params.query) search.set('query', params.query)
  if (params.hasGif) search.set('hasGif', 'true')
  if (params.limit) search.set('limit', String(params.limit))
  if (params.offset) search.set('offset', String(params.offset))
  const suffix = search.toString() ? `?${search.toString()}` : ''
  return request<ExerciseCatalogResponse>(`/v1/exercises${suffix}`)
}

export async function getCatalogExercise(datasetExerciseId: string): Promise<ExerciseCatalogItem> {
  const data = await request<{ exercise: ExerciseCatalogItem }>(`/v1/exercises/catalog/${encodeURIComponent(datasetExerciseId)}`)
  return data.exercise
}

async function request<T = unknown>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(`${apiBase}${path}`, {
    ...init,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...init.headers,
    },
  })
  if (!response.ok) {
    let message = response.statusText
    let code: string | undefined
    try {
      const body = await response.json()
      message = body.error?.message ?? message
      code = body.error?.code
    } catch {
      // Keep HTTP status text when the response is not JSON.
    }
    throw new ApiError(response.status, message, code)
  }
  if (response.status === 204) {
    return undefined as T
  }
  return response.json()
}
