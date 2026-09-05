import { cn } from '../app/cn'
import type { CheckpointStatus, ProgressCheckpoint, TrainingDay, TrainingRow } from '../app/api'

type TrainingDaysProps = {
  days: TrainingDay[]
  checkpoints: ProgressCheckpoint[]
  selectedExerciseKey: string
  onSelectExercise: (exerciseKey: string) => void
  onUpdateCheckpoint: (dayId: TrainingDay['id'], row: TrainingRow, status: CheckpointStatus, completedSets?: number) => Promise<void>
  isSaving: (dayId: TrainingDay['id'], exerciseKey: string) => boolean
}

const kindLabels: Record<string, string> = {
  main: 'Тяж.',
  light: 'Легк.',
  assistance: 'Подс.',
  gpp: 'ОФП',
}

const kgFormatter = new Intl.NumberFormat('ru-RU', {
  maximumFractionDigits: 1,
})

export function TrainingDays({ days, checkpoints, selectedExerciseKey, onSelectExercise, onUpdateCheckpoint, isSaving }: TrainingDaysProps) {
  return (
    <section aria-labelledby="training-days-title" className="min-w-0">
      <div className="mb-3 flex items-end justify-between gap-3">
        <div>
          <h2 id="training-days-title" className="text-balance text-base font-semibold text-slate-950">
            Тренировочные дни
          </h2>
          <p className="text-pretty text-sm text-slate-600">Считайте подходы кнопками +/− или отметьте упражнение целиком. День закроется после последнего упражнения.</p>
        </div>
      </div>
      <div className="grid gap-4">
        {days.map((day) => {
          const closed = dayDone(day, checkpoints)
          return (
            <article key={day.id} className="overflow-hidden rounded-2xl border border-slate-200 bg-white">
              <div className="flex items-center justify-between gap-3 border-b border-slate-200 bg-slate-50 px-4 py-2">
                <div>
                  <h3 className="text-sm font-semibold text-slate-950">{day.label}</h3>
                  <span className="text-xs font-medium text-slate-600">{day.focus}</span>
                </div>
                <span className={cn('shrink-0 text-xs font-semibold', closed ? 'text-emerald-700' : 'text-slate-500')}>
                  {closed ? 'День закрыт' : `${doneCount(day, checkpoints)}/${day.rows.length}`}
                </span>
              </div>

              <div className="lg:hidden">
                <div className="divide-y divide-slate-100">
                  {day.rows.map((row) => (
                    <div
                      key={row.rowId}
                      className={cn('p-4', row.exerciseKey === selectedExerciseKey && 'bg-sky-50')}
                    >
                      <div className="flex flex-col gap-4">
                        <div className="min-w-0">
                          <div className="flex items-start gap-2">
                            <span className="mt-3 shrink-0 rounded bg-slate-100 px-1.5 py-1 text-xs font-semibold text-slate-500">{kindLabels[row.kind]}</span>
                            <button
                              className="min-h-11 min-w-0 flex-1 py-2 text-left text-base font-semibold text-slate-950 underline-offset-4 hover:underline focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
                              type="button"
                              onClick={() => onSelectExercise(row.exerciseKey)}
                            >
                              {row.exerciseName}
                            </button>
                          </div>
                          <div className="mt-2 grid grid-cols-[1fr_1.3fr_.7fr] gap-2 rounded-xl bg-slate-50 p-3 text-sm">
                            <div>
                              <div className="text-slate-500">Подходы</div>
                              <div className="font-mono text-slate-800 tabular-nums">{row.prescription.setsRepsText}</div>
                            </div>
                            <div>
                              <div className="text-slate-500">Вес / RPE</div>
                              <div className="font-mono text-slate-950 tabular-nums">{loadText(row)}</div>
                            </div>
                            <div>
                              <div className="text-slate-500">Марк.</div>
                              <div className="text-slate-500">{row.prescription.unit ?? ''}</div>
                            </div>
                          </div>
                        </div>
                        <CheckpointControls
                          checkpoint={findCheckpoint(day.id, row.exerciseKey, checkpoints)}
                          disabled={isSaving(day.id, row.exerciseKey)}
                          exerciseName={row.exerciseName}
                          prescribedSets={row.prescription.sets}
                          onUpdate={(status, completedSets) => void onUpdateCheckpoint(day.id, row, status, completedSets)}
                        />
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              <div className="hidden overflow-x-auto lg:block" role="region" aria-labelledby={`${day.id}-caption`} tabIndex={0}>
                <table className="w-full min-w-[620px] border-collapse text-left text-sm">
                  <caption id={`${day.id}-caption`} className="sr-only">{day.label}: {day.focus}</caption>
                  <thead>
                    <tr className="border-b border-slate-200 text-xs uppercase text-slate-500">
                      <th className="w-16 px-3 py-2 font-semibold" scope="col">Тип</th>
                      <th className="px-3 py-2 font-semibold" scope="col">Упражнение</th>
                      <th className="w-28 px-3 py-2 font-semibold" scope="col">Подходы</th>
                      <th className="w-32 px-3 py-2 font-semibold" scope="col">Вес / RPE</th>
                      <th className="w-12 px-3 py-2 font-semibold" scope="col">Марк.</th>
                      <th className="w-44 px-3 py-2 font-semibold" scope="col">Прогресс</th>
                    </tr>
                  </thead>
                  <tbody>
                    {day.rows.map((row) => (
                      <tr
                        key={row.rowId}
                        className={cn('border-b border-slate-100 last:border-b-0', row.exerciseKey === selectedExerciseKey && 'bg-sky-50')}
                      >
                        <td className="px-3 py-2 text-xs font-semibold text-slate-500">{kindLabels[row.kind]}</td>
                        <th className="px-3 py-2 font-medium text-slate-950" scope="row">
                          <button
                            className="block w-full text-left underline-offset-4 hover:underline focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
                            type="button"
                            onClick={() => onSelectExercise(row.exerciseKey)}
                          >
                            {row.exerciseName}
                          </button>
                        </th>
                        <td className="px-3 py-2 font-mono text-slate-800 tabular-nums">{row.prescription.setsRepsText}</td>
                        <td className="px-3 py-2 font-mono text-slate-950 tabular-nums">{loadText(row)}</td>
                        <td className="px-3 py-2 text-xs text-slate-500">{row.prescription.unit ?? ''}</td>
                        <td className="px-3 py-2">
                          <CheckpointControls
                            checkpoint={findCheckpoint(day.id, row.exerciseKey, checkpoints)}
                            disabled={isSaving(day.id, row.exerciseKey)}
                            exerciseName={row.exerciseName}
                            prescribedSets={row.prescription.sets}
                            onUpdate={(status, completedSets) => void onUpdateCheckpoint(day.id, row, status, completedSets)}
                          />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </article>
          )
        })}
      </div>
    </section>
  )
}

function CheckpointControls({ checkpoint, disabled, exerciseName, prescribedSets, onUpdate }: {
  checkpoint?: ProgressCheckpoint
  disabled: boolean
  exerciseName: string
  prescribedSets?: number | null
  onUpdate: (status: CheckpointStatus, completedSets?: number) => void
}) {
  const completedSets = checkpoint?.completed?.sets ?? 0
  const hasSetCounter = typeof prescribedSets === 'number' && prescribedSets > 0
  const isDone = checkpoint?.status === 'done'
  const canIncrement = hasSetCounter && completedSets < prescribedSets
  const canDecrement = hasSetCounter && completedSets > 0

  function updateSetCount(nextSets: number) {
    const status = nextSets >= prescribedSets! ? 'done' : nextSets > 0 ? 'partial' : 'planned'
    onUpdate(status, nextSets)
  }

  return (
    <div className="flex flex-wrap items-center justify-between gap-2 lg:justify-end">
      {hasSetCounter ? (
        <div className="flex items-center gap-1" aria-label={`Подходы: ${completedSets} из ${prescribedSets}`}>
          <button
            aria-label={`Убавить подход: ${exerciseName}`}
            className="grid h-11 w-11 rounded-lg lg:h-8 lg:w-8 place-items-center border border-slate-300 bg-white text-lg font-medium leading-none text-slate-700 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
            disabled={disabled || !canDecrement}
            type="button"
            onClick={() => updateSetCount(completedSets - 1)}
          >
            −
          </button>
          <span className="min-w-12 text-center text-xs font-semibold tabular-nums text-slate-700">{completedSets}/{prescribedSets}</span>
          <button
            aria-label={`Добавить подход: ${exerciseName}`}
            className="grid h-11 w-11 rounded-lg lg:h-8 lg:w-8 place-items-center border border-slate-300 bg-white text-lg font-medium leading-none text-slate-700 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
            disabled={disabled || !canIncrement}
            type="button"
            onClick={() => updateSetCount(completedSets + 1)}
          >
            +
          </button>
        </div>
      ) : null}
      <button
        aria-label={isDone ? `Отменить выполнение: ${exerciseName}` : `Отметить выполненным: ${exerciseName}`}
        className={cn(
          'min-h-11 rounded-lg border px-4 text-sm font-semibold disabled:opacity-50 lg:min-h-8 lg:px-2 lg:text-xs transition focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2',
          isDone ? 'border-emerald-700 bg-emerald-50 text-emerald-800 hover:bg-emerald-100' : 'border-sky-700 bg-sky-700 text-white hover:bg-sky-800',
        )}
        disabled={disabled}
        type="button"
        onClick={() => onUpdate(isDone ? 'planned' : 'done', isDone ? completedSets || undefined : hasSetCounter ? prescribedSets : undefined)}
      >
        {isDone ? 'Отменить' : 'Сделано'}
      </button>
    </div>
  )
}

function findCheckpoint(dayId: TrainingDay['id'], exerciseKey: string, checkpoints: ProgressCheckpoint[]) {
  return checkpoints.find((checkpoint) => checkpoint.dayId === dayId && checkpoint.exerciseKey === exerciseKey)
}

function isDone(dayId: TrainingDay['id'], exerciseKey: string, checkpoints: ProgressCheckpoint[]) {
  return checkpoints.some((checkpoint) => checkpoint.dayId === dayId && checkpoint.exerciseKey === exerciseKey && checkpoint.status === 'done')
}

function doneCount(day: TrainingDay, checkpoints: ProgressCheckpoint[]) {
  return day.rows.filter((row) => isDone(day.id, row.exerciseKey, checkpoints)).length
}

function dayDone(day: TrainingDay, checkpoints: ProgressCheckpoint[]) {
  return day.rows.length > 0 && doneCount(day, checkpoints) === day.rows.length
}

function loadText(row: TrainingRow): string {
  if (row.prescription.weightText) return row.prescription.weightText
  if (typeof row.prescription.weightKg === 'number') return `${kgFormatter.format(row.prescription.weightKg)} кг`
  if (row.prescription.rpeText) return row.prescription.rpeText
  return ''
}
