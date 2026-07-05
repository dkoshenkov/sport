import { Modal } from './Modal'

type MarkInfoProps = {
  isOpen: boolean
  onClose: () => void
}

export function MarkInfo({ isOpen, onClose }: MarkInfoProps) {
  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Что такое Марк?">
      <div className="space-y-4 text-sm text-slate-700">
        <p>
          В контексте этой программы "Марк" обозначает тип нагрузки или категорию упражнения:
        </p>
        
        <ul className="list-disc space-y-1 pl-5">
          <li><strong>т</strong> — тяжелая нагрузка (основные упражнения)</li>
          <li><strong>л</strong> — легкая нагрузка (вспомогательные упражнения)</li>
          <li><strong>подс.</strong> — подсобные упражнения</li>
          <li><strong>ОФП</strong> — общая физическая подготовка</li>
        </ul>
        
        <p>
          Эти маркеры помогают быстро определить характер нагрузки в каждом упражнении 
          и понять его роль в тренировочном процессе.
        </p>
        
        <p>
          Например:
        </p>
        
        <ul className="list-disc space-y-1 pl-5">
          <li>Становая тяга — "т" (тяжелая нагрузка)</li>
          <li>Жим лежа — "л" (легкая нагрузка в некоторых днях)</li>
          <li>Гуд-морнинг — "подс." (подсобное упражнение)</li>
          <li>Подтягивания — "ОФП" (общая физическая подготовка)</li>
        </ul>
      </div>
    </Modal>
  )
}