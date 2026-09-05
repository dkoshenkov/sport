import { useEffect, useState } from 'react'
import { getCurrentCycle, listCurrentCycleProgress, upsertCurrentCycleCheckpoint, type ProgressCheckpoint, type ProgramCycle, type ProgramOptions, type TrainingPlan, type TrainingDay, type TrainingRow } from '../app/api'
import { ExerciseDetailsPanel } from './ExerciseDetailsPanel'
import { Modal } from './Modal'
import { MarkInfo } from './MarkInfo'
import { RpeInfo } from './RpeInfo'
import { SettingsPanel } from './SettingsPanel'
import { TrainingDays } from './TrainingDays'

type ProgramShellProps = {
  cycle: ProgramCycle
  options: ProgramOptions
  plan: TrainingPlan
  onCycleSaved: (cycle: ProgramCycle) => void
  onRefreshPlan: () => Promise<void>
  onReloadWorkspace: () => Promise<void>
}

export function ProgramShell({ cycle, options, plan, onCycleSaved, onRefreshPlan, onReloadWorkspace }: ProgramShellProps) {
  const firstExercise = plan.days[0]?.rows[0]?.exerciseKey ?? 'deadlift'
  const [selectedExerciseKey, setSelectedExerciseKey] = useState(firstExercise)
  const [isExerciseOpen, setIsExerciseOpen] = useState(false)
  const [isSettingsOpen, setIsSettingsOpen] = useState(false)
  const [isRpeInfoOpen, setIsRpeInfoOpen] = useState(false)
  const [isMarkInfoOpen, setIsMarkInfoOpen] = useState(false)
  const [checkpoints, setCheckpoints] = useState<ProgressCheckpoint[]>([])
  const [isProgressLoading, setIsProgressLoading] = useState(true)
  const [progressError, setProgressError] = useState('')
  const [savingKey, setSavingKey] = useState('')

  useEffect(() => {
    let cancelled = false
    setIsProgressLoading(true)
    setProgressError('')
    void listCurrentCycleProgress(cycle.currentWeek)
      .then((items) => {
        if (!cancelled) setCheckpoints(items)
      })
      .catch((error) => {
        if (!cancelled) setProgressError(error instanceof Error ? error.message : 'Не удалось загрузить чекпоинты.')
      })
      .finally(() => {
        if (!cancelled) setIsProgressLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [cycle.currentWeek, cycle.id])

  async function saved(cycle: ProgramCycle) {
    onCycleSaved(cycle)
    await onRefreshPlan()
  }

  async function updateCheckpoint(day: TrainingDay['id'], row: TrainingRow, status: ProgressCheckpoint['status'], completedSets?: number) {
    const key = `${day}:${row.exerciseKey}`
    if (savingKey) return
    setSavingKey(key)
    setProgressError('')
    try {
      const checkpoint = await upsertCurrentCycleCheckpoint({
        week: cycle.currentWeek,
        dayId: day,
        exerciseKey: row.exerciseKey,
        rowKind: row.kind,
        status,
        completed: completedSets === undefined ? undefined : { sets: completedSets },
      })
      setCheckpoints((items) => [...items.filter((item) => !(item.dayId === checkpoint.dayId && item.exerciseKey === checkpoint.exerciseKey)), checkpoint])
      const nextCycle = await getCurrentCycle()
      if (!nextCycle) {
        await onReloadWorkspace()
        return
      }
      await onCycleSaved(nextCycle)
    } catch (error) {
      setProgressError(error instanceof Error ? error.message : 'Не удалось сохранить чекпоинт.')
    } finally {
      setSavingKey('')
    }
  }

  return (
    <div className="text-slate-950">
      <a className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 focus:bg-slate-950 focus:px-3 focus:py-2 focus:text-sm focus:font-semibold focus:text-white" href="#main">
        Перейти к программе
      </a>
      <header className="mb-4 overflow-hidden rounded-2xl border border-slate-200 bg-white">
        <div className="flex flex-col gap-3 px-4 py-3 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h1 className="text-balance text-xl font-semibold text-slate-950">{cycle.title}</h1>
            <p className="mt-1 text-pretty text-sm text-slate-600">Активный тренировочный цикл</p>
          </div>
          <dl className="grid grid-cols-3 gap-3 border-t border-slate-100 pt-3 text-sm lg:border-0 lg:pt-0">
            <div>
              <dt className="text-xs font-medium text-slate-500">Неделя</dt>
              <dd className="font-semibold text-slate-950">{labelFor(options.weeks, cycle.currentWeek)}</dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-slate-500">Вариант</dt>
              <dd className="font-semibold text-slate-950">{labelFor(options.variants, cycle.settings.variant)}</dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-slate-500">Шаг</dt>
              <dd className="font-semibold text-slate-950">{labelFor(options.progressionSteps, cycle.settings.progressionStep)}</dd>
            </div>
          </dl>
        </div>
      </header>

      <main id="main" tabIndex={-1}>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-lg font-semibold text-slate-950">Тренировочный план</h2>
          <div className="grid w-full grid-cols-2 gap-2 sm:flex sm:w-auto sm:flex-wrap">
            <button
              className="h-10 border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
              type="button"
              onClick={() => setIsRpeInfoOpen(true)}
            >
              Что такое RPE?
            </button>
            <button
              className="h-10 border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
              type="button"
              onClick={() => setIsMarkInfoOpen(true)}
            >
              Что такое Марк?
            </button>
            <button
              className="col-span-2 h-10 border border-sky-700 bg-sky-700 px-3 text-sm font-semibold text-white transition hover:bg-sky-800 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
              type="button"
              onClick={() => setIsSettingsOpen(true)}
            >
              Настройки
            </button>
          </div>
        </div>

        {plan.warnings.length ? (
          <div className="mb-4 border-l-2 border-amber-500 bg-white px-3 py-2 text-sm text-slate-700">
            {plan.warnings.join(' ')}
          </div>
        ) : null}

        {progressError ? <p className="mb-4 border-l-2 border-red-600 bg-white px-3 py-2 text-sm text-red-800" role="alert">{progressError}</p> : null}
        {isProgressLoading ? <p className="mb-4 text-sm text-slate-600">Загрузка чекпоинтов…</p> : null}

        <div className="grid items-start gap-4 lg:grid-cols-[minmax(0,1fr)_340px]">
          <TrainingDays
            checkpoints={checkpoints}
            days={plan.days}
            isSaving={(day, exerciseKey) => savingKey === `${day}:${exerciseKey}`}
            onSelectExercise={(key) => {
              setSelectedExerciseKey(key)
              if (window.matchMedia('(max-width: 1023px)').matches) setIsExerciseOpen(true)
            }}
            onUpdateCheckpoint={updateCheckpoint}
            selectedExerciseKey={selectedExerciseKey}
          />
          <div className="hidden lg:sticky lg:top-4 lg:block">
            <ExerciseDetailsPanel exerciseKey={selectedExerciseKey} />
          </div>
        </div>
      </main>

      {isExerciseOpen && (
        <Modal isOpen onClose={() => setIsExerciseOpen(false)} title="Техника упражнения">
          <ExerciseDetailsPanel exerciseKey={selectedExerciseKey} />
        </Modal>
      )}
      <SettingsPanel
        cycle={cycle}
        isOpen={isSettingsOpen}
        options={options}
        onClose={() => setIsSettingsOpen(false)}
        onSaved={saved}
      />
      <RpeInfo isOpen={isRpeInfoOpen} onClose={() => setIsRpeInfoOpen(false)} />
      <MarkInfo isOpen={isMarkInfoOpen} onClose={() => setIsMarkInfoOpen(false)} />
    </div>
  )
}

function labelFor(options: Array<{ id: string; label: string }>, value: string) {
  return options.find((option) => option.id === value)?.label ?? value
}
