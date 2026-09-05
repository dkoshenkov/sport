import { exerciseLabel } from '../app/exerciseLabels'

export function ExerciseTags({ primary, secondary = [], equipment, bodyPart }: {
  primary: string[]; secondary?: string[]; equipment?: string | null; bodyPart?: string | null
}) {
  return <span className="flex flex-wrap gap-1.5">
    {primary.map((muscle) => <span key={`primary-${muscle}`} className="exercise-tag exercise-tag--primary" title="Основная мышца">{exerciseLabel(muscle)}</span>)}
    {secondary.filter((m) => !primary.includes(m)).map((muscle) => <span key={`secondary-${muscle}`} className="exercise-tag exercise-tag--secondary" title="Дополнительная мышца">{exerciseLabel(muscle)}</span>)}
    {equipment && <span className="exercise-tag exercise-tag--equipment" title="Оборудование">{exerciseLabel(equipment)}</span>}
    {bodyPart && <span className="exercise-tag" title="Часть тела">{exerciseLabel(bodyPart)}</span>}
  </span>
}
