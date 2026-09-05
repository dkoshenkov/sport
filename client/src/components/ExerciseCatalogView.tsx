import { useEffect, useState } from 'react'
import { listExercises, type ExerciseCatalogItem, type ExerciseCatalogResponse } from '../app/api'
import { cn } from '../app/cn'
import { Modal } from './Modal'
import { ExerciseTags } from './ExerciseTags'
import { exerciseLabel } from '../app/exerciseLabels'

const pageSize = 24
const emptyCatalog: ExerciseCatalogResponse = { items: [], total: 0, limit: pageSize, offset: 0, muscles: [], equipment: [], bodyParts: [] }

export function ExerciseCatalogView() {
  const [filters, setFilters] = useState({ query: '', muscle: '', equipment: '', bodyPart: '', hasImage: false })
  const [detailOpen, setDetailOpen] = useState(false)
  const [offset, setOffset] = useState(0)
  const [catalog, setCatalog] = useState(emptyCatalog)
  const [selected, setSelected] = useState<ExerciseCatalogItem | null>(null)
  const [completedRequest, setCompletedRequest] = useState('')
  const [error, setError] = useState('')
  const [retry, setRetry] = useState(0)
  const requestKey = JSON.stringify({ filters, offset, retry })
  const loading = completedRequest !== requestKey
  const activeFilters = Object.values(filters).filter(Boolean).length

  useEffect(() => {
    let cancelled = false
    const timer = window.setTimeout(() => {
      setError('')
      listExercises({ ...filters, query: filters.query.trim(), limit: pageSize, offset })
        .then((response) => {
          if (cancelled) return
          setCatalog(response)
          setSelected((current) => response.items.find((item) => item.datasetExerciseId === current?.datasetExerciseId) ?? response.items[0] ?? null)
        })
        .catch((err) => { if (!cancelled) setError(err instanceof Error ? err.message : 'Не удалось загрузить каталог.') })
        .finally(() => { if (!cancelled) setCompletedRequest(requestKey) })
    }, 250)
    return () => { cancelled = true; window.clearTimeout(timer) }
  }, [filters, offset, retry, requestKey])

  function updateFilter(key: keyof typeof filters, value: string | boolean) {
    setFilters((current) => ({ ...current, [key]: value }))
    setOffset(0)
    setSelected(null)
  }

  const detailsPanel = (
<aside className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm xl:sticky xl:top-5 xl:max-h-[calc(100dvh-2.5rem)] xl:overflow-y-auto" aria-label="Техника упражнения" aria-live="polite">
          {selected && !loading && !error ? <>
            <div className="border-b border-slate-100 p-5"><p className="text-xs font-bold uppercase tracking-widest text-sky-700">Техника упражнения</p><h2 className="mt-2 text-xl font-semibold text-slate-950">{selected.nameRu || selected.name}</h2>{selected.nameRu && <p className="mt-1 text-xs text-slate-500">{selected.name}</p>}</div>
            <ExerciseImage key={selected.datasetExerciseId} exercise={selected} />
            <div className="space-y-5 p-5">
              <div><h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-500">Основные мышцы</h3><ExerciseTags primary={selected.targetMuscles} /></div>
              {selected.secondaryMuscles.length > 0 && <div><h3 className="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-500">Также работают</h3><ExerciseTags primary={[]} secondary={selected.secondaryMuscles} /></div>}
              <ExerciseTags primary={[]} equipment={selected.equipment} bodyPart={selected.bodyPart} />
              <div><h3 className="mb-3 text-sm font-semibold text-slate-950">Как выполнять</h3>{selected.instructions.length ? <ol className="space-y-3">{selected.instructions.map((step, i) => <li key={`${i}-${step}`} className="flex gap-3 text-sm leading-6 text-slate-600"><span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-sky-50 text-xs font-bold text-sky-700">{i + 1}</span><span>{step}</span></li>)}</ol> : <p className="text-sm text-slate-500">Описание техники пока не добавлено.</p>}</div>
            </div>
          </> : <div className="px-6 py-16 text-center text-sm text-slate-500">{loading ? 'Загружаем упражнения…' : 'Выбери карточку, чтобы посмотреть технику и работающие мышцы.'}</div>}
        </aside>
  )

  return (
    <section aria-labelledby="catalog-title" className="space-y-6">
      <div className="catalog-hero">
        <div>
          <p className="text-xs font-bold uppercase tracking-[0.2em] text-sky-700">Библиотека движений</p>
          <h1 id="catalog-title" className="mt-2 text-3xl font-semibold tracking-tight text-slate-950 sm:text-4xl">Найди своё упражнение</h1>
          <p className="mt-3 max-w-xl text-sm leading-6 text-slate-600">Выбери мышечную группу и оборудование. Изучи технику и посмотри, какие мышцы включаются в работу.</p>
        </div>
        <div className="flex flex-wrap gap-2 text-xs" aria-label="Обозначения тегов">
          <span className="exercise-tag exercise-tag--primary">Основные мышцы</span>
          <span className="exercise-tag exercise-tag--secondary">Дополнительные</span>
          <span className="exercise-tag exercise-tag--equipment">Оборудование</span>
        </div>
      </div>

      <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm sm:p-5">
        <label htmlFor="exercise-search" className="text-sm font-semibold text-slate-800">Поиск упражнений</label>
        <input id="exercise-search" type="search" autoComplete="off" placeholder="Название, мышца или оборудование…" value={filters.query} onChange={(event) => updateFilter('query', event.target.value)} className="catalog-input mt-2 w-full" />
        <div className="mt-4 grid gap-3 sm:grid-cols-3">
          <Filter label="Мышцы" value={filters.muscle} values={catalog.muscles} onChange={(value) => updateFilter('muscle', value)} />
          <Filter label="Оборудование" value={filters.equipment} values={catalog.equipment} onChange={(value) => updateFilter('equipment', value)} />
          <Filter label="Часть тела" value={filters.bodyPart} values={catalog.bodyParts} onChange={(value) => updateFilter('bodyPart', value)} />
        </div>
        <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
          <label className="flex items-center gap-2 text-sm text-slate-600"><input type="checkbox" className="size-4 accent-sky-700" checked={filters.hasImage} onChange={(event) => updateFilter('hasImage', event.target.checked)} />С демонстрацией техники</label>
          {activeFilters > 0 && <button className="text-sm font-semibold text-sky-700 hover:underline" onClick={() => { setFilters({ query: '', muscle: '', equipment: '', bodyPart: '', hasImage: false }); setOffset(0); setSelected(null) }}>Сбросить фильтры · {activeFilters}</button>}
        </div>
      </div>

      {error && !loading && <div role="alert" className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-800">{error}<button className="ml-3 font-semibold underline" onClick={() => setRetry((value) => value + 1)}>Повторить</button></div>}

      <div className="grid items-start gap-6 xl:grid-cols-[minmax(0,1fr)_340px]">
        <div aria-busy={loading}>
          <div className="mb-4 flex items-center justify-between gap-2">
            <h2 className="text-base font-semibold text-slate-900">Упражнения</h2>
            <span className="text-sm text-slate-500" aria-live="polite">{loading ? 'Загрузка…' : `${catalog.total} найдено`}</span>
          </div>
          {loading ? <div className="grid grid-cols-2 gap-4 lg:grid-cols-3">{Array.from({ length: 6 }, (_, i) => <div key={i} className="h-72 animate-pulse rounded-2xl bg-slate-200 motion-reduce:animate-none" />)}</div> : !error && catalog.items.length === 0 ? (
            <div className="rounded-2xl border border-dashed border-slate-300 bg-white px-6 py-16 text-center"><h3 className="font-semibold text-slate-900">Ничего не найдено</h3><p className="mt-2 text-sm text-slate-500">Попробуй другое название или убери часть фильтров.</p></div>
          ) : !error && <div className="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-3">
            {catalog.items.map((exercise) => <button key={exercise.datasetExerciseId} type="button" aria-pressed={selected?.datasetExerciseId === exercise.datasetExerciseId} onClick={() => { setSelected(exercise); if (window.innerWidth < 1280) setDetailOpen(true) }} className={cn('exercise-card group', selected?.datasetExerciseId === exercise.datasetExerciseId && 'exercise-card--selected')}>
              <ExerciseImage key={exercise.datasetExerciseId} exercise={exercise} />
              <span className="block p-3 sm:p-4">
                <span className="mb-3 block min-h-10 text-sm font-semibold leading-5 text-slate-900">{exercise.nameRu || exercise.name}</span>
                <ExerciseTags primary={exercise.targetMuscles} secondary={exercise.secondaryMuscles} equipment={exercise.equipment} />
              </span>
            </button>)}
          </div>}
          {!loading && !error && catalog.total > pageSize && <nav aria-label="Страницы каталога" className="mt-5 flex items-center justify-between gap-2">
            <button className="catalog-page" disabled={offset === 0} onClick={() => { setOffset(Math.max(0, offset - pageSize)); setSelected(null) }}>← Назад</button>
            <span className="text-xs text-slate-500">{offset + 1}–{Math.min(offset + pageSize, catalog.total)} из {catalog.total}</span>
            <button className="catalog-page" disabled={offset + pageSize >= catalog.total} onClick={() => { setOffset(offset + pageSize); setSelected(null) }}>Далее →</button>
          </nav>}
        </div>
        <div className="hidden xl:sticky xl:top-5 xl:block">{detailsPanel}</div>
        {detailOpen && <Modal isOpen onClose={() => setDetailOpen(false)} title="Техника упражнения">{detailsPanel}</Modal>}
      </div>
    </section>
  )
}

function Filter({ label, value, values, onChange }: { label: string; value: string; values: string[]; onChange: (value: string) => void }) {
  return <label className="text-xs font-semibold text-slate-500">{label}<select aria-label={label} className="catalog-input mt-1.5 w-full" value={value} onChange={(event) => onChange(event.target.value)}><option value="">Все</option>{values.map((item) => <option key={item} value={item}>{exerciseLabel(item)}</option>)}</select></label>
}

function ExerciseImage({ exercise }: { exercise: ExerciseCatalogItem }) {
  const [failed, setFailed] = useState(false)
  return <span className="flex aspect-square items-center justify-center overflow-hidden bg-white p-3 text-center text-xs text-slate-400">
    {!failed && exercise.media.status === 'available' && exercise.media.imageUrl ? <img alt={`${exercise.nameRu || exercise.name}: демонстрация`} className="size-full object-contain" loading="lazy" src={exercise.media.imageUrl} width={320} height={320} onError={() => setFailed(true)} /> : <span className="px-4">Демонстрация пока недоступна</span>}
  </span>
}
