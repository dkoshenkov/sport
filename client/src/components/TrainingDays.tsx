import { cn } from '../app/cn'
import type { ProgressCheckpoint, TrainingDay, TrainingRow } from '../app/api'

type TrainingDaysProps = {
  days: TrainingDay[]
  checkpoints: ProgressCheckpoint[]
  selectedExerciseKey: string
  onSelectExercise: (exerciseKey: string) => void
  onToggleCheckpoint: (dayId: TrainingDay['id'], row: TrainingRow) => Promise<void>
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

export function TrainingDays({ days, checkpoints, selectedExerciseKey, onSelectExercise, onToggleCheckpoint, isSaving }: TrainingDaysProps) {
  return (
    <section aria-labelledby="training-days-title" className="min-w-0">
      <div className="mb-3 flex items-end justify-between gap-3">
        <div>
          <h2 id="training-days-title" className="text-balance text-base font-semibold text-slate-950">
            Тренировочные дни
          </h2>
          <p className="text-pretty text-sm text-slate-600">Отметьте каждое выполненное упражнение. День закроется после последней отметки.</p>
        </div>
      </div>
      <div className="grid gap-4">
        {days.map((day) => {
          const closed = dayDone(day, checkpoints)
          return (
            <article key={day.id} className="border border-slate-200 bg-white">
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
                      <div className="flex items-start justify-between gap-3">
                        <div className="flex-1">
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-semibold text-slate-500">{kindLabels[row.kind]}</span>
                            <button
                              className="text-left text-sm font-medium text-slate-950 underline-offset-4 hover:underline focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
                              type="button"
                              onClick={() => onSelectExercise(row.exerciseKey)}
                            >
                              {row.exerciseName}
                            </button>
                          </div>
                          <div className="mt-2 grid grid-cols-3 gap-2 text-xs">
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
                        <CheckpointToggle
                          checked={isDone(day.id, row.exerciseKey, checkpoints)}
                          disabled={isSaving(day.id, row.exerciseKey)}
                          exerciseName={row.exerciseName}
                          onChange={() => void onToggleCheckpoint(day.id, row)}
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
                      <th className="w-20 px-3 py-2 font-semibold" scope="col">Готово</th>
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
                          <CheckpointToggle
                            checked={isDone(day.id, row.exerciseKey, checkpoints)}
                            disabled={isSaving(day.id, row.exerciseKey)}
                            exerciseName={row.exerciseName}
                            onChange={() => void onToggleCheckpoint(day.id, row)}
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

function CheckpointToggle({ checked, disabled, exerciseName, onChange }: { checked: boolean; disabled: boolean; exerciseName: string; onChange: () => void }) {
  return (
    <label className="flex shrink-0 items-center gap-2 text-xs font-semibold text-slate-600">
      <input
        aria-label={checked ? `Снять отметку: ${exerciseName}` : `Отметить выполненным: ${exerciseName}`}
        checked={checked}
        className="h-5 w-5 accent-emerald-600 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
        disabled={disabled}
        type="checkbox"
        onChange={onChange}
      />
    </label>
  )
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
