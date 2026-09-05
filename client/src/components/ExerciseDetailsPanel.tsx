import { useEffect, useState } from 'react'
import { ExerciseTags } from './ExerciseTags'
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
    setDetails(null)
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
    <aside className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm lg:sticky lg:top-4" aria-labelledby="exercise-details-title">
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

        {!error && details?.media.status === 'available' && details.media.imageUrl ? (
          <>
            <div className="aspect-video w-full overflow-hidden border border-slate-200 bg-slate-50">
              <img
                alt={`${details.name}: демонстрация упражнения`}
                className="size-full object-contain"
                height={details.media.height ?? 240}
                loading="lazy"
                src={details.media.imageUrl}
                width={details.media.width ?? 320}
              />
            </div>
            <ExerciseMetadata details={details} />
          </>
        ) : null}

        {!error && details && details.media.status !== 'available' ? (
          <div className="space-y-3">
            <div className="grid aspect-video place-items-center border border-dashed border-slate-300 bg-slate-50 px-4 text-center">
              <p className="text-pretty text-sm text-slate-600">Изображение не подключено для этого упражнения.</p>
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
      <div className="space-y-4">
        <div><h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-500">Основные мышцы</h3><ExerciseTags primary={details.targetMuscles ?? []} /></div>
        {!!details.secondaryMuscles?.length && <div><h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-500">Также работают</h3><ExerciseTags primary={[]} secondary={details.secondaryMuscles} /></div>}
        <ExerciseTags primary={[]} equipment={details.equipment} />
      </div>
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
          Демонстрация может отличаться от варианта упражнения в программе.
        </p>
      ) : null}
    </>
  )
}
