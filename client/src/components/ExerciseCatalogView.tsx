import { useEffect, useState } from 'react'
import {
  getCatalogExercise,
  listExercises,
  type ExerciseCatalogItem,
} from '../app/api'
import { cn } from '../app/cn'

type CatalogState =
  | { status: 'idle' | 'loading'; items: ExerciseCatalogItem[]; total: number }
  | { status: 'ready'; items: ExerciseCatalogItem[]; total: number }
  | { status: 'error'; items: ExerciseCatalogItem[]; total: number; message: string }

type DetailState =
  | { status: 'idle' | 'loading' | 'missing'; exercise?: never; message?: never }
  | { status: 'ready'; exercise: ExerciseCatalogItem; message?: never }
  | { status: 'error'; exercise?: never; message: string }

export function ExerciseCatalogView() {
  const [query, setQuery] = useState('')
  const [hasGif, setHasGif] = useState(false)
  const [selectedId, setSelectedId] = useState('')
  const [catalog, setCatalog] = useState<CatalogState>({ status: 'idle', items: [], total: 0 })
  const [detail, setDetail] = useState<DetailState>({ status: 'idle' })

  useEffect(() => {
    let cancelled = false
    const handle = window.setTimeout(() => {
      setCatalog((current) => ({ status: 'loading', items: current.items, total: current.total }))
      listExercises({ query: query.trim(), hasGif, limit: 40 })
        .then((response) => {
          if (cancelled) return
          setCatalog({ status: 'ready', items: response.items, total: response.total })
          setSelectedId((current) => current || response.items[0]?.datasetExerciseId || '')
        })
        .catch((error) => {
          if (!cancelled) {
            setCatalog({
              status: 'error',
              items: [],
              total: 0,
              message: error instanceof Error ? error.message : 'Не удалось загрузить каталог.',
            })
          }
        })
    }, 250)
    return () => {
      cancelled = true
      window.clearTimeout(handle)
    }
  }, [query, hasGif])

  useEffect(() => {
    let cancelled = false
    if (!selectedId) {
      setDetail({ status: 'missing' })
      return () => {
        cancelled = true
      }
    }
    setDetail({ status: 'loading' })
    getCatalogExercise(selectedId)
      .then((exercise) => {
        if (!cancelled) setDetail({ status: 'ready', exercise })
      })
      .catch((error) => {
        if (!cancelled) {
          setDetail({ status: 'error', message: error instanceof Error ? error.message : 'Не удалось загрузить упражнение.' })
        }
      })
    return () => {
      cancelled = true
    }
  }, [selectedId])

  return (
    <section aria-labelledby="catalog-title">
      <div className="mb-4">
        <h1 id="catalog-title" className="text-xl font-semibold text-slate-950">Каталог упражнений</h1>
        <p className="mt-1 text-sm text-slate-600">Поиск по dataset, русским названиям программы, оборудованию и целевым мышцам.</p>
      </div>

      <div className="mb-4 border border-slate-200 bg-white p-4">
        <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-end">
          <label className="block text-sm font-medium text-slate-700" htmlFor="exercise-search">
            Поиск
            <input
              id="exercise-search"
              className="mt-1 h-10 w-full border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none transition focus:border-sky-700 focus:ring-2 focus:ring-sky-700/20"
              autoComplete="off"
              name="exercise-search"
              placeholder="reverse bench, жим обратным хватом..."
              type="search"
              value={query}
              onChange={(event) => {
                setQuery(event.target.value)
                setSelectedId('')
              }}
            />
          </label>
          <label className="flex min-h-10 items-center gap-2 text-sm font-medium text-slate-700">
            <input
              checked={hasGif}
              className="size-4 accent-sky-700"
              name="has-gif"
              type="checkbox"
              onChange={(event) => {
                setHasGif(event.target.checked)
                setSelectedId('')
              }}
            />
            Только с GIF
          </label>
        </div>
      </div>

      {catalog.status === 'error' ? (
        <p className="mb-4 border-l-2 border-red-600 bg-white px-3 py-2 text-sm text-red-700" role="alert">
          {catalog.message}
        </p>
      ) : null}

      <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_380px]">
        <ExerciseCatalogGrid
          items={catalog.items}
          isLoading={catalog.status === 'loading'}
          selectedId={selectedId}
          total={catalog.total}
          onSelect={setSelectedId}
        />
        <ExerciseCatalogDetails state={detail} />
      </div>
    </section>
  )
}

function ExerciseCatalogGrid({ items, selectedId, total, isLoading, onSelect }: {
  items: ExerciseCatalogItem[]
  selectedId: string
  total: number
  isLoading: boolean
  onSelect: (id: string) => void
}) {
  if (!items.length && !isLoading) {
    return (
      <div className="grid min-h-48 place-items-center border border-dashed border-slate-300 bg-white p-6 text-center">
        <p className="text-sm text-slate-600">Ничего не найдено. Измените запрос или отключите фильтр GIF.</p>
      </div>
    )
  }

  return (
    <div className="border border-slate-200 bg-white">
      <div className="sticky top-0 z-10 flex items-center justify-between border-b border-slate-200 bg-white/95 px-4 py-3">
        <h2 className="text-sm font-semibold text-slate-950">Упражнения</h2>
        <span className="text-xs text-slate-600" aria-live="polite">{isLoading ? 'Загрузка…' : `${total} найдено`}</span>
      </div>
      <div className="grid gap-3 p-3 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
        {items.map((exercise) => {
          const title = exercise.nameRu || exercise.name
          const hasMedia = exercise.media.status === 'available' && Boolean(exercise.media.gifUrl)
          return (
            <button
              key={exercise.datasetExerciseId}
              className={cn(
                'group overflow-hidden border border-slate-200 bg-white text-left shadow-sm transition hover:-translate-y-0.5 hover:border-sky-700 hover:shadow-md focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2',
                selectedId === exercise.datasetExerciseId && 'border-sky-700 bg-sky-50',
              )}
              type="button"
              onClick={() => onSelect(exercise.datasetExerciseId)}
            >
              <span className="flex aspect-[3/4] w-full items-center justify-center overflow-hidden border-b border-slate-200 bg-slate-100 text-center text-xs text-slate-500">
                {hasMedia ? (
                  <img
                    alt=""
                    className="size-full object-contain"
                    height={exercise.media.height ?? 240}
                    loading="lazy"
                    src={exercise.media.gifUrl ?? ''}
                    width={exercise.media.width ?? 180}
                  />
                ) : 'Нет GIF'}
              </span>
              <span className="block min-w-0 p-3">
                <span className="line-clamp-2 min-h-10 text-sm font-semibold leading-5 text-slate-950">{title}</span>
                <span className="mt-1 block truncate text-xs text-slate-600">{exercise.name}</span>
                <span className="mt-3 flex flex-wrap gap-1">
                  {exercise.bodyPart ? <Tag>{exercise.bodyPart}</Tag> : null}
                  {exercise.equipment ? <Tag>{exercise.equipment}</Tag> : null}
                  {exercise.targetMuscles[0] ? <Tag>{exercise.targetMuscles[0]}</Tag> : null}
                </span>
              </span>
            </button>
          )
        })}
      </div>
    </div>
  )
}

function Tag({ children }: { children: React.ReactNode }) {
  return (
    <span className="max-w-full truncate border border-slate-200 bg-slate-50 px-1.5 py-0.5 text-[11px] font-medium capitalize text-slate-600">
      {children}
    </span>
  )
}

function ExerciseCatalogDetails({ state }: { state: DetailState }) {
  const exercise = state.status === 'ready' ? state.exercise : null
  const title = exercise ? exercise.nameRu || exercise.name : 'Детали упражнения'
  const hasMedia = exercise?.media.status === 'available' && Boolean(exercise.media.gifUrl)

  return (
    <aside className="border border-slate-200 bg-white xl:sticky xl:top-4" aria-labelledby="catalog-detail-title">
      <div className="border-b border-slate-200 px-4 py-3">
        <h2 id="catalog-detail-title" className="text-base font-semibold text-slate-950">{title}</h2>
        <p className="mt-1 text-sm text-slate-600">{exercise?.datasetExerciseId ?? (state.status === 'loading' ? 'Загрузка…' : 'Выберите упражнение')}</p>
      </div>
      <div className="space-y-4 p-4">
        {state.status === 'error' ? (
          <div className="grid aspect-video place-items-center border border-dashed border-slate-300 bg-slate-50 px-4 text-center">
            <p className="text-sm text-slate-600" role="alert">{state.message}</p>
          </div>
        ) : null}

        {state.status === 'loading' ? (
          <div className="grid aspect-video place-items-center border border-dashed border-slate-300 bg-slate-50 px-4 text-center">
            <p className="text-sm text-slate-600">Загрузка деталей…</p>
          </div>
        ) : null}

        {exercise ? (
          <>
            <div className="grid aspect-[3/4] max-h-[420px] place-items-center overflow-hidden border border-slate-200 bg-slate-50">
              {hasMedia ? (
                <img
                  alt={`${title}: демонстрация упражнения`}
                  className="size-full object-contain"
                  height={exercise.media.height ?? 240}
                  loading="lazy"
                  src={exercise.media.gifUrl ?? ''}
                  width={exercise.media.width ?? 320}
                />
              ) : (
                <p className="px-4 text-center text-sm text-slate-600">GIF не подключен для этого упражнения.</p>
              )}
            </div>

            <dl className="grid grid-cols-[6rem_1fr] gap-x-3 gap-y-2 text-sm">
              <dt className="font-medium text-slate-600">Dataset</dt>
              <dd className="text-slate-950">{exercise.name}</dd>
              <dt className="font-medium text-slate-600">Оборуд.</dt>
              <dd className="text-slate-950">{exercise.equipment ?? 'Не указано'}</dd>
              <dt className="font-medium text-slate-600">Цель</dt>
              <dd className="text-slate-950">{exercise.targetMuscles.join(', ') || 'Не указано'}</dd>
              <dt className="font-medium text-slate-600">Доп.</dt>
              <dd className="text-slate-950">{exercise.secondaryMuscles.join(', ') || 'Не указано'}</dd>
            </dl>

            {exercise.instructions.length ? (
              <div>
                <h3 className="mb-2 text-sm font-semibold text-slate-950">Инструкция</h3>
                <ol className="list-decimal space-y-1 pl-5 text-sm leading-6 text-slate-700">
                  {exercise.instructions.map((instruction) => (
                    <li key={instruction}>{instruction}</li>
                  ))}
                </ol>
              </div>
            ) : null}
          </>
        ) : null}
      </div>
    </aside>
  )
}
