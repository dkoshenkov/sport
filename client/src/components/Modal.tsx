import { useEffect, useId, useRef } from 'react'
import { cn } from '../app/cn'

type ModalProps = {
  isOpen: boolean
  onClose: () => void
  title: string
  panelClassName?: string
  children: React.ReactNode
}

export function Modal({ isOpen, onClose, title, panelClassName = 'max-w-md sm:max-w-lg', children }: ModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const titleId = useId()

  useEffect(() => {
    if (isOpen) {
      if (!dialogRef.current?.open) dialogRef.current?.showModal()
    } else {
      dialogRef.current?.close()
    }
  }, [isOpen])

  const handleBackdropClick = (e: React.MouseEvent<HTMLDialogElement>) => {
    if (e.target === dialogRef.current) {
      onClose()
    }
  }

  return (
    <dialog
      ref={dialogRef}
      aria-labelledby={titleId}
      className="fixed inset-0 z-50 m-0 h-dvh w-screen overscroll-contain bg-black/50 p-0"
      onClick={handleBackdropClick}
      onClose={onClose}
    >
      <div
        className={cn(
          'absolute left-1/2 top-1/2 max-h-[90vh] w-[90vw] -translate-x-1/2 -translate-y-1/2 overflow-auto rounded-lg border border-slate-200 bg-white shadow-xl',
          panelClassName,
        )}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="border-b border-slate-200 px-4 py-3">
          <div className="flex items-start justify-between">
            <h2 id={titleId} className="text-balance text-base font-semibold text-slate-950">{title}</h2>
            <button
              className="text-slate-400 hover:text-slate-500 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
              type="button"
              onClick={onClose}
              aria-label="Закрыть"
            >
              <svg aria-hidden="true" className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
        <div className="p-4">
          {children}
        </div>
      </div>
    </dialog>
  )
}
