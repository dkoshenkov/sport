import { useEffect, useState } from 'react'
import {
  saveCycle,
  type CycleSettings,
  type ProgramCycle,
  type ProgramOptions,
  type ProgramWeek,
} from '../app/api'
import { CycleSettingsForm } from './CycleSettingsForm'
import { Modal } from './Modal'

type SettingsPanelProps = {
  cycle: ProgramCycle
  options: ProgramOptions
  isOpen: boolean
  onClose: () => void
  onSaved: (cycle: ProgramCycle) => Promise<void>
}

export function SettingsPanel({ cycle, options, isOpen, onClose, onSaved }: SettingsPanelProps) {
  const [title, setTitle] = useState(cycle.title)
  const [week, setWeek] = useState<ProgramWeek>(cycle.currentWeek)
  const [settings, setSettings] = useState<CycleSettings>(cycle.settings)
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (!isOpen) return
    setTitle(cycle.title)
    setWeek(cycle.currentWeek)
    setSettings(cycle.settings)
    setError('')
  }, [cycle, isOpen])

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    setIsSubmitting(true)
    try {
      const saved = await saveCycle(cycle.id, title.trim(), week, settings)
      await onSaved(saved)
      onClose()
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : 'Не удалось сохранить цикл.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Modal isOpen={isOpen} title="Настройки цикла" panelClassName="max-w-2xl" onClose={onClose}>
      <CycleSettingsForm
        currentWeek={week}
        error={error}
        isSubmitting={isSubmitting}
        options={options}
        settings={settings}
        submitLabel="Сохранить цикл"
        title={title}
        onSettingsChange={setSettings}
        onSubmit={submit}
        onTitleChange={setTitle}
        onWeekChange={setWeek}
      />
    </Modal>
  )
}
