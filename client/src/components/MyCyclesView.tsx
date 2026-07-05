import { useEffect, useState } from 'react'
import {
  activateCycle,
  createCycle,
  saveCycle,
  settingsFromProfile,
  type AthleteProfile,
  type CycleSettings,
  type ProgramCycle,
  type ProgramOptions,
  type ProgramWeek,
} from '../app/api'
import { CycleSettingsForm } from './CycleSettingsForm'
import { Modal } from './Modal'

type MyCyclesViewProps = {
  cycles: ProgramCycle[]
  profile: AthleteProfile
  options: ProgramOptions
  onChanged: () => Promise<void>
  onOpenProgram: () => void
}

type EditorState =
  | { mode: 'create'; cycle?: never }
  | { mode: 'edit'; cycle: ProgramCycle }

const statusLabels: Record<ProgramCycle['status'], string> = {
  active: 'Активный',
  completed: 'Завершён',
  archived: 'Архив',
}

export function MyCyclesView({ cycles, profile, options, onChanged, onOpenProgram }: MyCyclesViewProps) {
  const [editor, setEditor] = useState<EditorState | null>(null)
  const [pendingCycleId, setPendingCycleId] = useState('')
  const [error, setError] = useState('')

  async function makeActive(cycle: ProgramCycle, openAfter = false) {
    setError('')
    setPendingCycleId(cycle.id)
    try {
      if (cycle.status !== 'active') {
        await activateCycle(cycle.id)
        await onChanged()
      }
      if (openAfter) onOpenProgram()
    } catch (activationError) {
      setError(activationError instanceof Error ? activationError.message : 'Не удалось активировать цикл.')
    } finally {
      setPendingCycleId('')
    }
  }

  return (
    <section aria-labelledby="cycles-title">
      <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 id="cycles-title" className="text-xl font-semibold text-slate-950">Мои циклы</h1>
          <p className="mt-1 text-sm text-slate-600">История программ и выбор активного тренировочного цикла.</p>
        </div>
        <button
          className="h-10 border border-sky-700 bg-sky-700 px-3 text-sm font-semibold text-white transition hover:bg-sky-800 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
          type="button"
          onClick={() => setEditor({ mode: 'create' })}
        >
          Создать цикл
        </button>
      </div>

      {error ? (
        <p className="mb-4 border-l-2 border-red-600 bg-white px-3 py-2 text-sm text-red-700" role="alert">
          {error}
        </p>
      ) : null}

      {cycles.length === 0 ? (
        <div className="border border-dashed border-slate-300 bg-white p-6">
          <h2 className="text-base font-semibold text-slate-950">Циклов пока нет</h2>
          <p className="mt-1 text-sm text-slate-600">Создайте первый цикл из профильных значений 1ПМ и выбранных упражнений.</p>
        </div>
      ) : (
        <>
          <div className="grid gap-3 lg:hidden">
            {cycles.map((cycle) => (
              <CycleCard
                key={cycle.id}
                cycle={cycle}
                isPending={pendingCycleId === cycle.id}
                options={options}
                onActivate={() => void makeActive(cycle)}
                onEdit={() => setEditor({ mode: 'edit', cycle })}
                onOpen={() => void makeActive(cycle, true)}
              />
            ))}
          </div>

          <div className="hidden overflow-x-auto border border-slate-200 bg-white lg:block" role="region" aria-labelledby="cycles-table-caption" tabIndex={0}>
            <table className="w-full min-w-[760px] border-collapse text-left text-sm">
              <caption id="cycles-table-caption" className="sr-only">Мои тренировочные циклы</caption>
              <thead className="border-b border-slate-200 bg-slate-50 text-xs uppercase text-slate-500">
                <tr>
                  <th className="px-3 py-2 font-semibold" scope="col">Название</th>
                  <th className="w-28 px-3 py-2 font-semibold" scope="col">Статус</th>
                  <th className="w-28 px-3 py-2 font-semibold" scope="col">Неделя</th>
                  <th className="w-28 px-3 py-2 font-semibold" scope="col">Вариант</th>
                  <th className="w-24 px-3 py-2 font-semibold" scope="col">Шаг</th>
                  <th className="w-32 px-3 py-2 font-semibold" scope="col">Прогресс</th>
                  <th className="w-72 px-3 py-2 font-semibold" scope="col">Действия</th>
                </tr>
              </thead>
              <tbody>
                {cycles.map((cycle) => (
                  <tr key={cycle.id} className="border-b border-slate-100 last:border-b-0">
                    <th className="px-3 py-2 font-semibold text-slate-950" scope="row">{cycle.title}</th>
                    <td className="px-3 py-2 text-slate-700">{statusLabels[cycle.status]}</td>
                    <td className="px-3 py-2 text-slate-700">{labelFor(options.weeks, cycle.currentWeek)}</td>
                    <td className="px-3 py-2 text-slate-700">{labelFor(options.variants, cycle.settings.variant)}</td>
                    <td className="px-3 py-2 text-slate-700">{labelFor(options.progressionSteps, cycle.settings.progressionStep)}</td>
                    <td className="px-3 py-2 font-mono text-slate-700 tabular-nums">{progressText(cycle)}</td>
                    <td className="px-3 py-2">
                      <div className="flex flex-wrap gap-2">
                        <button className="h-9 border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2" type="button" onClick={() => void makeActive(cycle, true)}>
                          Открыть
                        </button>
                        {cycle.status !== 'active' ? (
                          <button
                            className="h-9 border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60"
                            disabled={pendingCycleId === cycle.id}
                            type="button"
                            onClick={() => void makeActive(cycle)}
                          >
                            {pendingCycleId === cycle.id ? 'Активация…' : 'Сделать активным'}
                          </button>
                        ) : null}
                        <button className="h-9 border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2" type="button" onClick={() => setEditor({ mode: 'edit', cycle })}>
                          Настроить
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}

      <CycleEditorModal
        editor={editor}
        options={options}
        profile={profile}
        onChanged={onChanged}
        onClose={() => setEditor(null)}
      />
    </section>
  )
}

function CycleCard({ cycle, options, isPending, onOpen, onActivate, onEdit }: {
  cycle: ProgramCycle
  options: ProgramOptions
  isPending: boolean
  onOpen: () => void
  onActivate: () => void
  onEdit: () => void
}) {
  return (
    <article className="border border-slate-200 bg-white p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold text-slate-950">{cycle.title}</h2>
          <p className="mt-1 text-sm text-slate-600">{statusLabels[cycle.status]} · {labelFor(options.weeks, cycle.currentWeek)}</p>
        </div>
        <span className="text-sm font-mono text-slate-600 tabular-nums">{progressText(cycle)}</span>
      </div>
      <dl className="mt-3 grid grid-cols-2 gap-2 text-sm">
        <div>
          <dt className="text-xs text-slate-500">Вариант</dt>
          <dd className="font-medium text-slate-800">{labelFor(options.variants, cycle.settings.variant)}</dd>
        </div>
        <div>
          <dt className="text-xs text-slate-500">Шаг</dt>
          <dd className="font-medium text-slate-800">{labelFor(options.progressionSteps, cycle.settings.progressionStep)}</dd>
        </div>
      </dl>
      <div className="mt-4 flex flex-wrap gap-2">
        <button className="h-10 border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2" type="button" onClick={onOpen}>
          Открыть
        </button>
        {cycle.status !== 'active' ? (
          <button className="h-10 border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60" disabled={isPending} type="button" onClick={onActivate}>
            {isPending ? 'Активация…' : 'Сделать активным'}
          </button>
        ) : null}
        <button className="h-10 border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2" type="button" onClick={onEdit}>
          Настроить
        </button>
      </div>
    </article>
  )
}

function CycleEditorModal({ editor, options, profile, onChanged, onClose }: {
  editor: EditorState | null
  options: ProgramOptions
  profile: AthleteProfile
  onChanged: () => Promise<void>
  onClose: () => void
}) {
  const [title, setTitle] = useState('Линейный цикл')
  const [week, setWeek] = useState<ProgramWeek>('week_1')
  const [settings, setSettings] = useState<CycleSettings>(() => settingsFromProfile(profile))
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (!editor) return
    setTitle(editor.mode === 'edit' ? editor.cycle.title : 'Линейный цикл')
    setWeek(editor.mode === 'edit' ? editor.cycle.currentWeek : 'week_1')
    setSettings(editor.mode === 'edit' ? editor.cycle.settings : settingsFromProfile(profile))
    setError('')
  }, [editor, profile])

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!editor) return
    setError('')
    setIsSubmitting(true)
    try {
      if (editor.mode === 'create') {
        await createCycle(title.trim(), 'week_1', settings)
      } else {
        await saveCycle(editor.cycle.id, title.trim(), week, settings)
      }
      await onChanged()
      onClose()
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : 'Не удалось сохранить цикл.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Modal isOpen={Boolean(editor)} title={editor?.mode === 'edit' ? 'Настройки цикла' : 'Создать цикл'} panelClassName="max-w-2xl" onClose={onClose}>
      <CycleSettingsForm
        currentWeek={week}
        error={error}
        isSubmitting={isSubmitting}
        options={options}
        settings={settings}
        showWeek={editor?.mode === 'edit'}
        submitLabel={editor?.mode === 'edit' ? 'Сохранить цикл' : 'Создать цикл'}
        title={title}
        onSettingsChange={setSettings}
        onSubmit={submit}
        onTitleChange={setTitle}
        onWeekChange={setWeek}
      />
    </Modal>
  )
}

function progressText(cycle: ProgramCycle) {
  const { done, partial, skipped, planned } = cycle.progressSummary
  const total = done + partial + skipped + planned
  return `${done}/${total}`
}

function labelFor(options: SelectOptionLike[], value: string) {
  return options.find((option) => option.id === value)?.label ?? value
}

type SelectOptionLike = {
  id: string
  label: string
}
