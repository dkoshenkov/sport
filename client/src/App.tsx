import { useCallback, useEffect, useState } from 'react'
import {
  ApiError,
  bootstrap,
  getAuthSession,
  getCurrentCyclePlan,
  getProfile,
  getProgramOptions,
  listCycles,
  type AthleteProfile,
  type CyclesResponse,
  type ProgramCycle,
  type ProgramOptions,
  type TrainingPlan,
  type User,
} from './app/api'
import { AuthGate } from './components/AuthGate'
import { WorkspaceShell } from './components/WorkspaceShell'

type WorkspaceTab = 'program' | 'cycles' | 'catalog'

function App() {
  const [user, setUser] = useState<User | null>(null)
  const [profile, setProfile] = useState<AthleteProfile | null>(null)
  const [options, setOptions] = useState<ProgramOptions | null>(null)
  const [cycles, setCycles] = useState<ProgramCycle[]>([])
  const [cycle, setCycle] = useState<ProgramCycle | null>(null)
  const [plan, setPlan] = useState<TrainingPlan | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState('')
  const [activeTab, setActiveTab] = useState<WorkspaceTab>('program')

  const clearSession = useCallback(() => {
    setUser(null)
    setProfile(null)
    setOptions(null)
    setCycles([])
    setCycle(null)
    setPlan(null)
    setError('')
    setActiveTab('program')
  }, [])

  const refreshPlan = useCallback(async () => {
    const nextPlan = await getCurrentCyclePlan()
    setPlan(nextPlan)
  }, [])

  const loadWorkspace = useCallback(async () => {
    const [nextProfile, nextOptions, cycleList] = await Promise.all([
      getProfile(),
      getProgramOptions(),
      listCycles(),
    ])
    const nextCycle = activeCycleFromList(cycleList)
    setProfile(nextProfile)
    setOptions(nextOptions)
    setCycles(cycleList.cycles)
    setCycle(nextCycle)
    if (nextCycle) {
      setPlan(await getCurrentCyclePlan())
    } else {
      setPlan(null)
    }
  }, [])

  const checkSession = useCallback(async () => {
    setIsLoading(true)
    setError('')
    try {
      await bootstrap()
      const session = await getAuthSession()
      if (!session.authenticated || !session.user) {
        clearSession()
        return
      }
      const nextUser = session.user
      setUser(nextUser)
      await loadWorkspace()
    } catch (loadError) {
      if (loadError instanceof ApiError && loadError.status === 401) {
        clearSession()
        return
      }
      setError(loadError instanceof Error ? loadError.message : 'Не удалось загрузить приложение.')
    } finally {
      setIsLoading(false)
    }
  }, [clearSession, loadWorkspace])

  useEffect(() => {
    void checkSession()
  }, [checkSession])

  async function authenticated(nextUser: User) {
    setUser(nextUser)
    setIsLoading(true)
    setError('')
    try {
      await loadWorkspace()
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : 'Не удалось загрузить данные аккаунта.')
    } finally {
      setIsLoading(false)
    }
  }

  async function cycleSaved(nextCycle: ProgramCycle) {
    setCycle(nextCycle)
    setCycles((current) => current.map((item) => item.id === nextCycle.id ? nextCycle : item))
    setPlan(await getCurrentCyclePlan())
  }

  async function reloadWorkspace() {
    setIsLoading(true)
    setError('')
    try {
      await loadWorkspace()
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : 'Не удалось загрузить данные аккаунта.')
    } finally {
      setIsLoading(false)
    }
  }

  if (isLoading) {
    return (
      <main className="grid min-h-dvh place-items-center bg-slate-100 px-4 text-slate-950">
        <div className="border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700">Загрузка программы…</div>
      </main>
    )
  }

  if (!user) {
    return <AuthGate onAuthenticated={authenticated} />
  }

  if (error) {
    return (
      <main className="grid min-h-dvh place-items-center bg-slate-100 px-4 text-slate-950">
        <section className="w-full max-w-md border border-slate-200 bg-white p-4">
          <h1 className="text-balance text-lg font-semibold text-slate-950">Не удалось открыть программу</h1>
          <p className="mt-2 text-pretty text-sm text-slate-600" role="alert">{error}</p>
          <button
            className="mt-4 h-10 border border-sky-700 bg-sky-700 px-3 text-sm font-semibold text-white transition hover:bg-sky-800 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
            type="button"
            onClick={() => void checkSession()}
          >
            Повторить
          </button>
        </section>
      </main>
    )
  }

  if (!profile || !options) {
    return null
  }

  return (
    <WorkspaceShell
      activeTab={activeTab}
      cycle={cycle}
      cycles={cycles}
      options={options}
      plan={plan}
      profile={profile}
      user={user}
      onCycleSaved={cycleSaved}
      onReloadWorkspace={reloadWorkspace}
      onRefreshPlan={refreshPlan}
      onSetTab={setActiveTab}
      onSignedOut={clearSession}
    />
  )
}

function activeCycleFromList(response: CyclesResponse) {
  return response.cycles.find((item) => item.id === response.currentCycleId)
    ?? response.cycles.find((item) => item.status === 'active')
    ?? null
}

export default App
