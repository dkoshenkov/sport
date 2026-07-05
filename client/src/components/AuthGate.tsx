import { useMemo, useState } from 'react'
import { login, register, type User } from '../app/api'

type AuthGateProps = {
  onAuthenticated: (user: User) => void
}

type AuthMode = 'login' | 'register'

export function AuthGate({ onAuthenticated }: AuthGateProps) {
  const [mode, setMode] = useState<AuthMode>('login')
  const [nickname, setNickname] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const isValid = useMemo(() => nickname.trim().length >= 3 && password.length >= 8, [nickname, password])

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!isValid || isSubmitting) return
    setError('')
    setIsSubmitting(true)
    try {
      const user = mode === 'login'
        ? await login(nickname.trim(), password)
        : await register(nickname.trim(), password)
      onAuthenticated(user)
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : 'Не удалось войти.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <main className="grid min-h-dvh place-items-center bg-slate-100 px-4 py-8 text-slate-950">
      <section className="w-full max-w-sm border border-slate-200 bg-white">
        <div className="border-b border-slate-200 px-4 py-3">
          <h1 className="text-balance text-lg font-semibold text-slate-950">Вход в программу</h1>
          <p className="mt-1 text-pretty text-sm text-slate-600">Доступ к циклу открыт только после входа.</p>
        </div>
        <form className="space-y-4 p-4" onSubmit={submit}>
          <div>
            <label className="block text-sm font-medium text-slate-700" htmlFor="nickname">
              Никнейм
            </label>
            <input
              id="nickname"
              autoComplete="username"
              className="mt-1 h-10 w-full border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none transition focus:border-sky-700 focus:ring-2 focus:ring-sky-700/20"
              minLength={3}
              name="nickname"
              required
              spellCheck={false}
              type="text"
              value={nickname}
              onChange={(event) => setNickname(event.target.value)}
            />
            <p className="mt-1 text-xs leading-5 text-slate-600">Минимум 3 символа: буквы, цифры, дефис или подчёркивание.</p>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700" htmlFor="password">
              Пароль
            </label>
            <input
              id="password"
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
              className="mt-1 h-10 w-full border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none transition focus:border-sky-700 focus:ring-2 focus:ring-sky-700/20"
              minLength={8}
              name="password"
              required
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />
            <p className="mt-1 text-xs leading-5 text-slate-600">Минимум 8 символов.</p>
          </div>

          {error ? (
            <p className="border-l-2 border-red-600 pl-3 text-sm text-red-700" role="alert">
              {error}
            </p>
          ) : null}

          <button
            className="h-10 w-full border border-sky-700 bg-sky-700 px-3 text-sm font-semibold text-white transition hover:bg-sky-800 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2 disabled:cursor-not-allowed disabled:border-slate-300 disabled:bg-slate-200 disabled:text-slate-500"
            disabled={!isValid || isSubmitting}
            type="submit"
          >
            {isSubmitting ? 'Отправка…' : mode === 'login' ? 'Войти' : 'Создать аккаунт'}
          </button>

          <button
            className="h-10 w-full border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-sky-700 focus:ring-offset-2"
            type="button"
            onClick={() => {
              setMode(mode === 'login' ? 'register' : 'login')
              setError('')
            }}
          >
            {mode === 'login' ? 'Нужен новый аккаунт' : 'Уже есть аккаунт'}
          </button>
        </form>
      </section>
    </main>
  )
}
