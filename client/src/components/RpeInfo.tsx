import { Modal } from './Modal'

type RpeInfoProps = {
  isOpen: boolean
  onClose: () => void
}

export function RpeInfo({ isOpen, onClose }: RpeInfoProps) {
  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Что такое RPE?">
      <div className="space-y-4 text-sm text-slate-700">
        <p>
          RPE (Rate of Perceived Exertion) — это субъективная шкала оценки_intensity_ нагрузки во время выполнения упражнения.
        </p>
        
        <p>
          В пауэрлифтинге используется шкала от 1 до 10, где:
        </p>
        
        <ul className="list-disc space-y-1 pl-5">
          <li><strong>1-2</strong> — минимальная нагрузка</li>
          <li><strong>3-4</strong> — легкая нагрузка</li>
          <li><strong>5-6</strong> — умеренная нагрузка</li>
          <li><strong>7-8</strong> — тяжелая нагрузка</li>
          <li><strong>9-10</strong> — максимальная нагрузка</li>
        </ul>
        
        <p>
          Например, RPE 7 означает, что вы можете сделать еще 3 повторения с тем же весом, 
          а RPE 8 — что вы можете сделать еще 2 повторения.
        </p>
        
        <p>
          В этой программе RPE используется для регулирования нагрузки в подсобных упражнениях, 
          позволяя адаптировать тренировку под ваше текущее состояние.
        </p>
      </div>
    </Modal>
  )
}