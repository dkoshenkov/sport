import {
  benchAssistanceOptions,
  deadliftAssistanceOptions,
  gppAbsOptions,
  gppBicepsOptions,
  gppHorizontalPullOptions,
  gppOverheadPressOptions,
  gppTricepsOptions,
  gppVerticalPullOptions,
  squatAssistanceOptions,
  variants,
  weeks,
} from '../data/programOptions'
import type { LiftId, ProgramSelection } from '../domain/types'

type SettingsPanelProps = {
  selection: ProgramSelection
  onChange: (selection: ProgramSelection) => void
  onReset: () => void
}

type SelectOption = {
  id: string
  label: string
  note?: string
}

const liftLabels: Record<LiftId, string> = {
  deadlift: 'Становая 1ПМ',
  bench: 'Жим 1ПМ',
  squat: 'Присед 1ПМ',
}

export function SettingsPanel({ selection, onChange, onReset }: SettingsPanelProps) {
  function updateSelection<K extends keyof ProgramSelection>(key: K, value: ProgramSelection[K]) {
    onChange({ ...selection, [key]: value })
  }

  function updateOneRepMax(lift: LiftId, value: string) {
    const parsed = Number(value.replace(',', '.'))
    onChange({
      ...selection,
      oneRepMax: {
        ...selection.oneRepMax,
        [lift]: Number.isFinite(parsed) && parsed > 0 ? parsed : 0,
      },
    })
  }

  return (
    <aside className="border border-slate-200 bg-white">
      <div className="border-b border-slate-200 px-4 py-3">
        <h2 className="text-balance text-base font-semibold text-slate-950">Настройки цикла</h2>
        <p className="mt-1 text-pretty text-sm text-slate-600">Значения сохраняются локально на этом устройстве.</p>
      </div>

      <div className="space-y-5 p-4">
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
                min="0"
                name={`${lift}-1rm`}
                type="text"
                value={selection.oneRepMax[lift as LiftId] || ''}
                onChange={(event) => updateOneRepMax(lift as LiftId, event.target.value)}
              />
            </label>
          ))}
        </fieldset>

        <fieldset className="grid gap-3">
          <legend className="text-sm font-semibold text-slate-800">Неделя и вариант</legend>
          <SelectField
            id="week"
            label="Неделя"
            options={weeks}
            value={selection.week}
            onChange={(value) => updateSelection('week', value as ProgramSelection['week'])}
          />
          <SelectField
            id="variant"
            label="Вариант"
            options={variants}
            value={selection.variant}
            onChange={(value) => updateSelection('variant', value as ProgramSelection['variant'])}
          />
          <p className="text-pretty border-l-2 border-sky-700 pl-3 text-xs leading-5 text-slate-600">
            Шаг прогрессии зафиксирован как 4% для совместимости с XLSX.
          </p>
        </fieldset>

        <fieldset className="grid gap-3">
          <legend className="text-sm font-semibold text-slate-800">Подсобные упражнения</legend>
          <SelectField
            id="deadlift-assistance"
            label="К становой"
            options={deadliftAssistanceOptions}
            value={selection.deadliftAssistance}
            onChange={(value) => updateSelection('deadliftAssistance', value)}
          />
          <SelectField
            id="bench-assistance"
            label="К жиму"
            options={benchAssistanceOptions}
            value={selection.benchAssistance}
            onChange={(value) => updateSelection('benchAssistance', value)}
          />
          <SelectField
            id="squat-assistance"
            label="К приседу"
            options={squatAssistanceOptions}
            value={selection.squatAssistance}
            onChange={(value) => updateSelection('squatAssistance', value)}
          />
        </fieldset>

        <fieldset className="grid gap-3">
          <legend className="text-sm font-semibold text-slate-800">ОФП</legend>
          <SelectField
            id="horizontal-pull"
            label="Горизонтальная тяга"
            options={gppHorizontalPullOptions}
            value={selection.gppHorizontalPull}
            onChange={(value) => updateSelection('gppHorizontalPull', value)}
          />
          <SelectField
            id="biceps"
            label="Бицепс"
            options={gppBicepsOptions}
            value={selection.gppBiceps}
            onChange={(value) => updateSelection('gppBiceps', value)}
          />
          <SelectField
            id="abs"
            label="Пресс"
            options={gppAbsOptions}
            value={selection.gppAbs}
            onChange={(value) => updateSelection('gppAbs', value)}
          />
          <SelectField
            id="triceps"
            label="Трицепс"
            options={gppTricepsOptions}
            value={selection.gppTriceps}
            onChange={(value) => updateSelection('gppTriceps', value)}
          />
          <SelectField
            id="vertical-pull"
            label="Вертикальная тяга"
            options={gppVerticalPullOptions}
            value={selection.gppVerticalPull}
            onChange={(value) => updateSelection('gppVerticalPull', value)}
          />
          <SelectField
            id="overhead-press"
            label="Жим над головой"
            options={gppOverheadPressOptions}
            value={selection.gppOverheadPress}
            onChange={(value) => updateSelection('gppOverheadPress', value)}
          />
        </fieldset>

        <button
          className="h-10 w-full border border-slate-300 bg-slate-50 px-3 text-sm font-semibold text-slate-800 transition hover:bg-slate-100 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
          type="button"
          onClick={onReset}
        >
          Сбросить настройки
        </button>
      </div>
    </aside>
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
          <option key={option.id} value={option.id}>
            {option.note ? `${option.label} - ${option.note}` : option.label}
          </option>
        ))}
      </select>
    </label>
  )
}
