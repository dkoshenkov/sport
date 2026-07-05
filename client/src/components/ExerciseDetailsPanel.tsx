import { useEffect, useState } from 'react'
import { getExerciseDetails, type ExerciseDetails } from '../app/api'

type ExerciseDetailsPanelProps = {
  exerciseKey: string
  onUnauthorized?: () => void
}

export function ExerciseDetailsPanel({ exerciseKey, onUnauthorized }: ExerciseDetailsPanelProps) {
  const [details, setDetails] = useState<ExerciseDetails | null>(null)
  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    let cancelled = false
    setIsLoading(true)
    setError('')
    getExerciseDetails(exerciseKey)
      .then((exercise) => {
        if (!cancelled) setDetails(exercise)
      })
      .catch((loadError) => {
        if (cancelled) return
        if (loadError instanceof Error && 'status' in loadError && loadError.status === 401) {
          onUnauthorized?.()
          return
        }
        setDetails(null)
        setError(loadError instanceof Error ? loadError.message : 'Не удалось загрузить упражнение.')
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [exerciseKey, onUnauthorized])

  return (
    <aside className="border border-slate-200 bg-white lg:sticky lg:top-4" aria-labelledby="exercise-details-title">
      <div className="border-b border-slate-200 px-4 py-3">
        <h2 id="exercise-details-title" className="text-balance text-base font-semibold text-slate-950">
          Детали упражнения
        </h2>
        <p className="mt-1 text-pretty text-sm text-slate-600">
          {details?.name ?? (isLoading ? 'Загрузка…' : 'Выберите упражнение в таблице')}
        </p>
      </div>

      <div className="space-y-4 p-4">
        {error ? (
          <div className="grid aspect-video place-items-center border border-dashed border-slate-300 bg-slate-50 px-4 text-center">
            <p className="text-pretty text-sm text-slate-600" role="alert">{error}</p>
          </div>
        ) : null}

        {!error && details?.media.status === 'available' && details.media.gifUrl ? (
          <>
            <div className="aspect-video w-full overflow-hidden border border-slate-200 bg-slate-50">
              <img
                alt={`${details.name}: демонстрация упражнения`}
                className="size-full object-contain"
                height={details.media.height ?? 240}
                loading="lazy"
                src={details.media.gifUrl}
                width={details.media.width ?? 320}
              />
            </div>
            <ExerciseMetadata details={details} />
          </>
        ) : null}

        {!error && details && details.media.status !== 'available' ? (
          <div className="space-y-3">
            <div className="grid aspect-video place-items-center border border-dashed border-slate-300 bg-slate-50 px-4 text-center">
              <p className="text-pretty text-sm text-slate-600">GIF не подключен для этого упражнения.</p>
            </div>
            <ExerciseMetadata details={details} />
          </div>
        ) : null}

        {!error && !details ? (
          <div className="grid aspect-video place-items-center border border-dashed border-slate-300 bg-slate-50 px-4 text-center">
            <p className="text-pretty text-sm text-slate-600">{isLoading ? 'Загрузка деталей…' : 'Нет данных.'}</p>
          </div>
        ) : null}
      </div>
    </aside>
  )
}

function ExerciseMetadata({ details }: { details: ExerciseDetails }) {
  return (
    <>
      <dl className="grid grid-cols-[6rem_1fr] gap-x-3 gap-y-2 text-sm">
        <dt className="font-medium text-slate-600">Dataset</dt>
        <dd className="text-slate-950">{details.datasetName ?? details.datasetExerciseId ?? 'Нет соответствия'}</dd>
        <dt className="font-medium text-slate-600">Оборуд.</dt>
        <dd className="text-slate-950">{details.equipment ?? 'Не указано'}</dd>
        <dt className="font-medium text-slate-600">Цель</dt>
        <dd className="text-slate-950">{details.targetMuscles?.join(', ') || 'Не указано'}</dd>
        <dt className="font-medium text-slate-600">Доп.</dt>
        <dd className="text-slate-950">{details.secondaryMuscles?.join(', ') || 'Не указано'}</dd>
      </dl>
      {details.instructions?.length ? (
        <div>
          <h3 className="mb-2 text-sm font-semibold text-slate-950">Инструкция</h3>
          <ol className="list-decimal space-y-1 pl-5 text-sm leading-6 text-slate-700">
            {details.instructions.map((instruction) => (
              <li key={instruction}>{instruction}</li>
            ))}
          </ol>
        </div>
      ) : null}
      {details.aliasStatus !== 'confirmed' ? (
        <p className="text-pretty border-l-2 border-amber-500 pl-3 text-xs leading-5 text-slate-600">
          Соответствие с dataset требует проверки.
        </p>
      ) : null}
    </>
  )
}
