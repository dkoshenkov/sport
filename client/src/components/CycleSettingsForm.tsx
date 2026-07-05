import { useEffect, useState } from 'react'
import { getExerciseDetails, type CycleSettings, type ExerciseDetails, type LiftId, type ProgramOptions, type ProgramWeek, type SelectOption } from '../app/api'

type CycleSettingsFormProps = {
  title: string
  currentWeek: ProgramWeek
  settings: CycleSettings
  options: ProgramOptions
  submitLabel: string
  showWeek?: boolean
  isSubmitting?: boolean
  error?: string
  onTitleChange: (title: string) => void
  onWeekChange: (week: ProgramWeek) => void
  onSettingsChange: (settings: CycleSettings) => void
  onSubmit: (event: React.FormEvent<HTMLFormElement>) => void
}

const liftLabels: Record<LiftId, string> = {
  deadlift: 'Становая 1ПМ',
  bench: 'Жим 1ПМ',
  squat: 'Присед 1ПМ',
}

export function CycleSettingsForm({
  title,
  currentWeek,
  settings,
  options,
  submitLabel,
  showWeek = true,
  isSubmitting = false,
  error = '',
  onTitleChange,
  onWeekChange,
  onSettingsChange,
  onSubmit,
}: CycleSettingsFormProps) {
  const isValid = title.trim().length > 0
    && settings.oneRepMaxKg.deadlift > 0
    && settings.oneRepMaxKg.bench > 0
    && settings.oneRepMaxKg.squat > 0

  function updateOneRepMax(lift: LiftId, value: string) {
    const parsed = Number(value.replace(',', '.'))
    onSettingsChange({
      ...settings,
      oneRepMaxKg: {
        ...settings.oneRepMaxKg,
        [lift]: Number.isFinite(parsed) && parsed > 0 ? parsed : 0,
      },
    })
  }

  function updateSettings<K extends keyof CycleSettings>(key: K, value: CycleSettings[K]) {
    onSettingsChange({ ...settings, [key]: value })
  }

  return (
    <form className="space-y-5" onSubmit={onSubmit}>
      <fieldset className="grid gap-3">
        <legend className="text-sm font-semibold text-slate-800">Цикл</legend>
        <label className="block text-sm font-medium text-slate-700" htmlFor="cycle-title">
          Название
          <input
            id="cycle-title"
            autoComplete="off"
            className="mt-1 h-10 w-full border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none transition focus:border-sky-700 focus:ring-2 focus:ring-sky-700/20"
            name="cycle-title"
            required
            type="text"
            value={title}
            onChange={(event) => onTitleChange(event.target.value)}
          />
        </label>
        {showWeek ? (
          <SelectField
            id="cycle-week"
            label="Текущая неделя"
            options={options.weeks}
            value={currentWeek}
            onChange={(value) => onWeekChange(value as ProgramWeek)}
          />
        ) : null}
      </fieldset>

      <fieldset className="space-y-3">
        <legend className="text-sm font-semibold text-slate-800">Основные 1ПМ, кг</legend>
        {Object.entries(liftLabels).map(([lift, label]) => (
          <label key={lift} className="block text-sm font-medium text-slate-700" htmlFor={`${lift}-1rm`}>
            {label}
            <input
              id={`${lift}-1rm`}
              autoComplete="off"
              className="mt-1 h-10 w-full border border-slate-300 bg-white px-3 text-sm text-slate-950 tabular-nums outline-none transition focus:border-sky-700 focus:ring-2 focus:ring-sky-700/20"
              inputMode="decimal"
              min="1"
              name={`${lift}-1rm`}
              required
              type="text"
              value={settings.oneRepMaxKg[lift as LiftId] || ''}
              onChange={(event) => updateOneRepMax(lift as LiftId, event.target.value)}
            />
          </label>
        ))}
      </fieldset>

      <fieldset className="grid gap-3">
        <legend className="text-sm font-semibold text-slate-800">Вариант и шаг</legend>
        <SelectField
          id="variant"
          label="Вариант"
          options={options.variants}
          value={settings.variant}
          onChange={(value) => updateSettings('variant', value as CycleSettings['variant'])}
        />
        <SelectField
          id="progression-step"
          label="Шаг прогрессии"
          options={options.progressionSteps}
          value={settings.progressionStep}
          onChange={(value) => updateSettings('progressionStep', value as CycleSettings['progressionStep'])}
        />
      </fieldset>

      <fieldset className="grid gap-3">
        <legend className="text-sm font-semibold text-slate-800">Подсобные упражнения</legend>
        <AssistanceSelect
          id="deadlift-assistance"
          label="К становой"
          options={options.assistance.deadlift}
          value={settings.assistance.deadlift}
          onChange={(value) => updateSettings('assistance', { ...settings.assistance, deadlift: value })}
        />
        <AssistanceSelect
          id="bench-assistance"
          label="К жиму"
          options={options.assistance.bench}
          value={settings.assistance.bench}
          onChange={(value) => updateSettings('assistance', { ...settings.assistance, bench: value })}
        />
        <AssistanceSelect
          id="squat-assistance"
          label="К приседу"
          options={options.assistance.squat}
          value={settings.assistance.squat}
          onChange={(value) => updateSettings('assistance', { ...settings.assistance, squat: value })}
        />
      </fieldset>

      <fieldset className="grid gap-3">
        <legend className="text-sm font-semibold text-slate-800">ОФП</legend>
        <SelectField
          id="horizontal-pull"
          label="Горизонтальная тяга"
          options={options.gpp.horizontalPull}
          value={settings.gpp.horizontalPull ?? ''}
          onChange={(value) => updateSettings('gpp', { ...settings.gpp, horizontalPull: value || null })}
        />
        <SelectField
          id="biceps"
          label="Бицепс"
          options={options.gpp.biceps}
          value={settings.gpp.biceps ?? ''}
          onChange={(value) => updateSettings('gpp', { ...settings.gpp, biceps: value || null })}
        />
        <SelectField
          id="abs"
          label="Пресс"
          options={options.gpp.abs}
          value={settings.gpp.abs ?? ''}
          onChange={(value) => updateSettings('gpp', { ...settings.gpp, abs: value || null })}
        />
        <SelectField
          id="triceps"
          label="Трицепс"
          options={options.gpp.triceps}
          value={settings.gpp.triceps ?? ''}
          onChange={(value) => updateSettings('gpp', { ...settings.gpp, triceps: value || null })}
        />
        <PreviewSelectField
          id="vertical-pull"
          label="Вертикальная тяга"
          options={options.gpp.verticalPull}
          value={settings.gpp.verticalPull ?? ''}
          onChange={(value) => updateSettings('gpp', { ...settings.gpp, verticalPull: value || null })}
        />
        <PreviewSelectField
          id="overhead-press"
          label="Жим над головой"
          options={options.gpp.overheadPress}
          value={settings.gpp.overheadPress ?? ''}
          onChange={(value) => updateSettings('gpp', { ...settings.gpp, overheadPress: value || null })}
        />
      </fieldset>

      {error ? (
        <p className="border-l-2 border-red-600 pl-3 text-sm text-red-700" role="alert">
          {error}
        </p>
      ) : null}

      <button
        className="h-10 w-full border border-sky-700 bg-sky-700 px-3 text-sm font-semibold text-white transition hover:bg-sky-800 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2 disabled:cursor-not-allowed disabled:border-slate-300 disabled:bg-slate-200 disabled:text-slate-500"
        disabled={!isValid || isSubmitting}
        type="submit"
      >
        {isSubmitting ? 'Сохранение…' : submitLabel}
      </button>
    </form>
  )
}

function AssistanceSelect({ id, label, options, value, onChange }: {
  id: string
  label: string
  options: SelectOption[]
  value: string
  onChange: (value: string) => void
}) {
  return (
    <div className="grid gap-2">
      <SelectField id={id} label={label} options={options} value={value} onChange={onChange} />
      <ExerciseMiniPreview exerciseKey={value} />
    </div>
  )
}

function PreviewSelectField({ id, label, options, value, onChange }: {
  id: string
  label: string
  options: SelectOption[]
  value: string
  onChange: (value: string) => void
}) {
  return (
    <div className="grid gap-2">
      <SelectField id={id} label={label} options={options} value={value} onChange={onChange} />
      <ExerciseMiniPreview exerciseKey={value} />
    </div>
  )
}

type ExercisePreviewState =
  | { status: 'idle' | 'loading' | 'missing' | 'error'; details?: never }
  | { status: 'ready'; details: ExerciseDetails }

function ExerciseMiniPreview({ exerciseKey }: { exerciseKey: string }) {
  const [state, setState] = useState<ExercisePreviewState>({ status: 'idle' })

  useEffect(() => {
    let isCurrent = true

    if (!exerciseKey) {
      setState({ status: 'missing' })
      return () => {
        isCurrent = false
      }
    }

    setState({ status: 'loading' })
    getExerciseDetails(exerciseKey)
      .then((details) => {
        if (isCurrent) setState({ status: 'ready', details })
      })
      .catch(() => {
        if (isCurrent) setState({ status: 'error' })
      })

    return () => {
      isCurrent = false
    }
  }, [exerciseKey])

  const details = state.status === 'ready' ? state.details : null
  const media = details?.media
  const hasGif = media?.status === 'available' && Boolean(media.gifUrl)
  const title = details?.name ?? 'Упражнение'
  const subtitle = details?.targetMuscles?.length
    ? details.targetMuscles.slice(0, 3).join(', ')
    : details?.equipment ?? 'Данные упражнения'
  const statusText = state.status === 'loading'
    ? 'Загрузка GIF'
    : hasGif
      ? 'GIF подключен'
      : state.status === 'error'
        ? 'GIF недоступен'
        : 'GIF не подключен'

  return (
    <div className="grid grid-cols-[6rem_minmax(0,1fr)] gap-3 border border-slate-200 bg-slate-50 p-2">
      <div className="flex h-20 w-24 items-center justify-center overflow-hidden border border-slate-200 bg-white text-center text-[11px] leading-4 text-slate-500">
        {hasGif ? (
          <img
            alt={title}
            className="h-full w-full object-contain"
            height={media?.height ?? 120}
            loading="lazy"
            src={media?.gifUrl ?? ''}
            width={media?.width ?? 160}
          />
        ) : (
          <span className="px-2">{state.status === 'loading' ? 'Загрузка…' : 'Нет GIF'}</span>
        )}
      </div>
      <div className="min-w-0 self-center" aria-live="polite">
        <p className="truncate text-xs font-semibold text-slate-800">{title}</p>
        <p className="mt-1 truncate text-xs text-slate-600">{subtitle}</p>
        <p className="mt-2 text-xs text-slate-500">{statusText}</p>
      </div>
    </div>
  )
}

function SelectField({ id, label, options, value, onChange }: {
  id: string
  label: string
  options: SelectOption[]
  value: string
  onChange: (value: string) => void
}) {
  return (
    <label className="block text-sm font-medium text-slate-700" htmlFor={id}>
      {label}
      <select
        id={id}
        className="mt-1 h-10 w-full border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none transition focus:border-sky-700 focus:ring-2 focus:ring-sky-700/20"
        name={id}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        {options.map((option) => (
          <option key={option.id || 'skip'} value={option.id}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  )
}
