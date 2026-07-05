import { logout, type AthleteProfile, type ProgramCycle, type ProgramOptions, type TrainingPlan, type User } from '../app/api'
import { ExerciseCatalogView } from './ExerciseCatalogView'
import { MyCyclesView } from './MyCyclesView'
import { ProgramShell } from './ProgramShell'

type WorkspaceTab = 'program' | 'cycles' | 'catalog'

type WorkspaceShellProps = {
  activeTab: WorkspaceTab
  cycle: ProgramCycle | null
  cycles: ProgramCycle[]
  options: ProgramOptions
  plan: TrainingPlan | null
  profile: AthleteProfile
  user: User
  onCycleSaved: (cycle: ProgramCycle) => void
  onRefreshPlan: () => Promise<void>
  onReloadWorkspace: () => Promise<void>
  onSetTab: (tab: WorkspaceTab) => void
  onSignedOut: () => void
}

const tabs: Array<{ id: WorkspaceTab; label: string }> = [
  { id: 'program', label: 'Программа' },
  { id: 'cycles', label: 'Мои циклы' },
  { id: 'catalog', label: 'Каталог' },
]

export function WorkspaceShell({
  activeTab,
  cycle,
  cycles,
  options,
  plan,
  profile,
  user,
  onCycleSaved,
  onRefreshPlan,
  onReloadWorkspace,
  onSetTab,
  onSignedOut,
}: WorkspaceShellProps) {
  async function signOut() {
    try {
      await logout()
    } finally {
      onSignedOut()
    }
  }

  return (
    <div className="min-h-dvh bg-slate-100 text-slate-950">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto max-w-[1480px] px-4 py-3 sm:px-6">
          <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
            <div>
              <p className="text-xs font-semibold uppercase text-slate-500">Training program</p>
              <h1 className="mt-1 text-xl font-semibold text-slate-950">Рабочая программа</h1>
            </div>
            <div className="flex items-center justify-between gap-3 lg:justify-end">
              <span className="text-sm font-medium text-slate-700">{user.nickname}</span>
              <button
                className="h-10 border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
                type="button"
                onClick={signOut}
              >
                Выйти
              </button>
            </div>
          </div>

          <nav className="mt-4 flex gap-1 border-b border-slate-200" aria-label="Разделы приложения">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                className={`h-10 border-b-2 px-3 text-sm font-semibold transition focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2 ${
                  activeTab === tab.id
                    ? 'border-sky-700 text-sky-800'
                    : 'border-transparent text-slate-600 hover:text-slate-950'
                }`}
                aria-current={activeTab === tab.id ? 'page' : undefined}
                type="button"
                onClick={() => onSetTab(tab.id)}
              >
                {tab.label}
              </button>
            ))}
          </nav>
        </div>
      </header>

      <main className="mx-auto max-w-[1480px] px-4 py-4 sm:px-6">
        {activeTab === 'program' ? (
          cycle && plan ? (
            <ProgramShell
              cycle={cycle}
              options={options}
              plan={plan}
              onCycleSaved={onCycleSaved}
              onRefreshPlan={onRefreshPlan}
            />
          ) : (
            <NoActiveCycle onCreate={() => onSetTab('cycles')} />
          )
        ) : null}

        {activeTab === 'cycles' ? (
          <MyCyclesView
            cycles={cycles}
            options={options}
            profile={profile}
            onChanged={onReloadWorkspace}
            onOpenProgram={() => onSetTab('program')}
          />
        ) : null}

        {activeTab === 'catalog' ? <ExerciseCatalogView /> : null}
      </main>
    </div>
  )
}

function NoActiveCycle({ onCreate }: { onCreate: () => void }) {
  return (
    <section className="border border-dashed border-slate-300 bg-white p-6" aria-labelledby="no-active-cycle-title">
      <h2 id="no-active-cycle-title" className="text-lg font-semibold text-slate-950">Активного цикла нет</h2>
      <p className="mt-1 max-w-2xl text-sm text-slate-600">
        Можно смотреть каталог упражнений и историю циклов. Чтобы открыть тренировочный план, создайте новый цикл или сделайте один из существующих активным.
      </p>
      <button
        className="mt-4 h-10 border border-sky-700 bg-sky-700 px-3 text-sm font-semibold text-white transition hover:bg-sky-800 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
        type="button"
        onClick={onCreate}
      >
        Перейти к циклам
      </button>
    </section>
  )
}
