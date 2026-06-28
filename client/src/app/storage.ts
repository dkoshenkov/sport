import { defaultSelection } from '../data/programOptions'
import type { ProgramSelection } from '../domain/types'

const storageKey = 'linear-cycle-program-selection'

export function loadSelection(): ProgramSelection {
  try {
    const raw = window.localStorage.getItem(storageKey)
    if (!raw) return defaultSelection
    return { ...defaultSelection, ...JSON.parse(raw) }
  } catch {
    return defaultSelection
  }
}

export function saveSelection(selection: ProgramSelection) {
  window.localStorage.setItem(storageKey, JSON.stringify(selection))
}

export function resetStoredSelection() {
  window.localStorage.removeItem(storageKey)
}
