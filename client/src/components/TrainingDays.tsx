import { cn } from '../app/cn'
import type { TrainingDay } from '../domain/types'

type TrainingDaysProps = {
  days: TrainingDay[]
  selectedExerciseKey: string
  onSelectExercise: (exerciseKey: string) => void
}

const kindLabels: Record<string, string> = {
  main: 'Тяж.',
  light: 'Легк.',
  assistance: 'Подс.',
  gpp: 'ОФП',
}

export function TrainingDays({ days, selectedExerciseKey, onSelectExercise }: TrainingDaysProps) {
  return (
    <section aria-labelledby="training-days-title" className="min-w-0">
      <div className="mb-3 flex items-end justify-between gap-3">
        <div>
          <h2 id="training-days-title" className="text-balance text-base font-semibold text-slate-950">
            Тренировочные дни
          </h2>
          <p className="text-pretty text-sm text-slate-600">Клик по упражнению открывает детали справа.</p>
        </div>
      </div>
      <div className="grid gap-4">
        {days.map((day) => (
          <article key={day.id} className="border border-slate-200 bg-white">
            <div className="flex items-center justify-between border-b border-slate-200 bg-slate-50 px-4 py-2">
              <h3 className="text-sm font-semibold text-slate-950">{day.title}</h3>
              <span className="text-xs font-medium text-slate-600">{day.focus}</span>
            </div>
            <div
              aria-labelledby={`${day.id}-caption`}
              className="overflow-x-auto"
              role="region"
              tabIndex={0}
            >
              <table className="w-full min-w-[560px] border-collapse text-left text-sm">
                <caption id={`${day.id}-caption`} className="sr-only">
                  {day.title}: {day.focus}
                </caption>
                <thead>
                  <tr className="border-b border-slate-200 text-xs uppercase text-slate-500">
                    <th className="w-16 px-3 py-2 font-semibold" scope="col">
                      Тип
                    </th>
                    <th className="px-3 py-2 font-semibold" scope="col">
                      Упражнение
                    </th>
                    <th className="w-28 px-3 py-2 font-semibold" scope="col">
                      Подходы
                    </th>
                    <th className="w-32 px-3 py-2 font-semibold" scope="col">
                      Вес / RPE
                    </th>
                    <th className="w-12 px-3 py-2 font-semibold" scope="col">
                      Марк.
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {day.rows.map((row) => (
                    <tr
                      key={row.id}
                      className={cn(
                        'border-b border-slate-100 last:border-b-0',
                        row.exerciseKey === selectedExerciseKey && 'bg-sky-50',
                      )}
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
                      <td className="px-3 py-2 font-mono text-slate-800 tabular-nums">{row.prescription}</td>
                      <td className="px-3 py-2 font-mono text-slate-950 tabular-nums">{row.load}</td>
                      <td className="px-3 py-2 text-xs text-slate-500">{row.unit ?? ''}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </article>
        ))}
      </div>
    </section>
  )
}
