import { exerciseAliases } from '../data/exerciseAliases'
import { datasetGifUrl, exerciseDetailsByDatasetId } from '../data/exerciseDetails'

type ExerciseDetailsPanelProps = {
  exerciseKey: string
}

export function ExerciseDetailsPanel({ exerciseKey }: ExerciseDetailsPanelProps) {
  const alias = exerciseAliases[exerciseKey]
  const detail = alias?.datasetExerciseId ? exerciseDetailsByDatasetId[alias.datasetExerciseId] : null

  return (
    <aside className="border border-slate-200 bg-white" aria-labelledby="exercise-details-title">
      <div className="border-b border-slate-200 px-4 py-3">
        <h2 id="exercise-details-title" className="text-balance text-base font-semibold text-slate-950">
          Детали упражнения
        </h2>
        <p className="mt-1 text-pretty text-sm text-slate-600">{alias?.programNameRu ?? 'Выберите упражнение в таблице'}</p>
      </div>

      <div className="space-y-4 p-4">
        {detail ? (
          <>
            <div className="aspect-video w-full overflow-hidden border border-slate-200 bg-slate-50">
              <img
                alt={`${alias.programNameRu}: демонстрация упражнения`}
                className="size-full object-contain"
                height="240"
                loading="lazy"
                src={datasetGifUrl(detail.gifPath)}
                width="320"
              />
            </div>
            <dl className="grid grid-cols-[6rem_1fr] gap-x-3 gap-y-2 text-sm">
              <dt className="font-medium text-slate-600">Dataset</dt>
              <dd className="text-slate-950">{detail.datasetName}</dd>
              <dt className="font-medium text-slate-600">Оборуд.</dt>
              <dd className="text-slate-950">{detail.equipment}</dd>
              <dt className="font-medium text-slate-600">Цель</dt>
              <dd className="text-slate-950">{detail.target}</dd>
              <dt className="font-medium text-slate-600">Доп.</dt>
              <dd className="text-slate-950">{detail.secondary.join(', ')}</dd>
            </dl>
            <div>
              <h3 className="mb-2 text-sm font-semibold text-slate-950">Инструкция</h3>
              <ol className="list-decimal space-y-1 pl-5 text-sm leading-6 text-slate-700">
                {detail.instructions.map((instruction) => (
                  <li key={instruction}>{instruction}</li>
                ))}
              </ol>
            </div>
            {alias.notes ? <p className="text-pretty border-l-2 border-amber-500 pl-3 text-xs leading-5 text-slate-600">{alias.notes}</p> : null}
          </>
        ) : (
          <div className="space-y-3">
            <div className="grid aspect-video place-items-center border border-dashed border-slate-300 bg-slate-50 px-4 text-center">
              <p className="text-pretty text-sm text-slate-600">GIF не подключен для этого упражнения.</p>
            </div>
            <dl className="grid gap-2 text-sm">
              <div>
                <dt className="font-medium text-slate-600">Статус</dt>
                <dd className="text-slate-950">{alias?.status === 'needs_review' ? 'Нужно подтвердить соответствие' : 'Нет подтвержденного соответствия'}</dd>
              </div>
              {alias?.notes ? (
                <div>
                  <dt className="font-medium text-slate-600">Причина</dt>
                  <dd className="text-pretty text-slate-950">{alias.notes}</dd>
                </div>
              ) : null}
            </dl>
          </div>
        )}
      </div>
    </aside>
  )
}
