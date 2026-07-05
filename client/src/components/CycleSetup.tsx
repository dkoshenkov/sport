import { useState } from 'react'
import {
  saveCurrentCycle,
  settingsFromProfile,
  type AthleteProfile,
  type CycleSettings,
  type ProgramCycle,
  type ProgramOptions,
  type ProgramWeek,
} from '../app/api'
import { CycleSettingsForm } from './CycleSettingsForm'

type CycleSetupProps = {
  profile: AthleteProfile
  options: ProgramOptions
  onCycleCreated: (cycle: ProgramCycle) => Promise<void>
}

export function CycleSetup({ profile, options, onCycleCreated }: CycleSetupProps) {
  const [title, setTitle] = useState('Линейный цикл')
  const [week, setWeek] = useState<ProgramWeek>('week_1')
  const [settings, setSettings] = useState<CycleSettings>(() => settingsFromProfile(profile))
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    setIsSubmitting(true)
    try {
      const cycle = await saveCurrentCycle(title.trim(), week, settings)
      await onCycleCreated(cycle)
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : 'Не удалось создать цикл.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <main className="mx-auto max-w-2xl px-4 py-6 sm:px-6">
      <section className="border border-slate-200 bg-white">
        <div className="border-b border-slate-200 px-4 py-3">
          <h1 className="text-balance text-lg font-semibold text-slate-950">Создать активный цикл</h1>
          <p className="mt-1 text-pretty text-sm text-slate-600">Настройки сохранятся снимком цикла и не будут меняться от правок профиля.</p>
        </div>
        <div className="p-4">
          <CycleSettingsForm
            currentWeek={week}
            error={error}
            isSubmitting={isSubmitting}
            options={options}
            settings={settings}
            submitLabel="Создать цикл"
            title={title}
            onSettingsChange={setSettings}
            onSubmit={submit}
            onTitleChange={setTitle}
            onWeekChange={setWeek}
          />
        </div>
      </section>
    </main>
  )
}
