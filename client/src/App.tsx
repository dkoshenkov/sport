import { useEffect, useMemo, useState } from 'react'
import { ExerciseDetailsPanel } from './components/ExerciseDetailsPanel'
import { SettingsPanel } from './components/SettingsPanel'
import { TrainingDays } from './components/TrainingDays'
import { defaultSelection } from './data/programOptions'
import { calculateTrainingPlan, variantLabel, weekLabel } from './domain/program'
import type { ProgramSelection } from './domain/types'
import { loadSelection, resetStoredSelection, saveSelection } from './app/storage'

function App() {
  const [selection, setSelection] = useState<ProgramSelection>(() => loadSelection())
  const [selectedExerciseKey, setSelectedExerciseKey] = useState('deadlift')
  const plan = useMemo(() => calculateTrainingPlan(selection), [selection])

  useEffect(() => {
    document.cookie = 'init=1; Path=/; SameSite=Lax; Max-Age=31536000'
  }, [])

  useEffect(() => {
    saveSelection(selection)
  }, [selection])

  function resetSelection() {
    if (!window.confirm('Сбросить локальные настройки цикла?')) return
    resetStoredSelection()
    setSelection(defaultSelection)
    setSelectedExerciseKey('deadlift')
  }

  return (
    <div className="min-h-dvh bg-slate-100 text-slate-950">
      <a className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 focus:bg-slate-950 focus:px-3 focus:py-2 focus:text-sm focus:font-semibold focus:text-white" href="#main">
        Перейти к программе
      </a>
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-[1480px] flex-col gap-3 px-4 py-4 sm:px-6 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h1 className="text-balance text-xl font-semibold text-slate-950 sm:text-2xl">Линейный цикл для продвинутых</h1>
            <p className="mt-1 text-pretty text-sm text-slate-600">MVP-интерфейс тренировочной таблицы из XLSX</p>
          </div>
          <dl className="grid grid-cols-2 gap-x-6 gap-y-1 text-sm sm:grid-cols-4">
            <div>
              <dt className="text-xs font-medium text-slate-500">Неделя</dt>
              <dd className="font-semibold text-slate-950">{weekLabel(selection.week)}</dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-slate-500">Вариант</dt>
              <dd className="font-semibold text-slate-950">{variantLabel(selection.variant)}</dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-slate-500">Шаг</dt>
              <dd className="font-semibold text-slate-950">4%</dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-slate-500">Сессия</dt>
              <dd className="font-semibold text-slate-950">init=1</dd>
            </div>
          </dl>
        </div>
      </header>

      <main id="main" className="mx-auto grid max-w-[1480px] gap-4 px-4 py-4 sm:px-6 lg:grid-cols-[320px_minmax(0,1fr)_340px]" tabIndex={-1}>
        <SettingsPanel selection={selection} onChange={setSelection} onReset={resetSelection} />
        <TrainingDays days={plan} selectedExerciseKey={selectedExerciseKey} onSelectExercise={setSelectedExerciseKey} />
        <ExerciseDetailsPanel exerciseKey={selectedExerciseKey} />
      </main>
    </div>
  )
}

export default App
