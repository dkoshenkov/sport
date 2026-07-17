const strengthMarks = ['💪', '🏋️', '🦾', '🏋️‍♂️', '🏆', '⚡', '🔥']

function isStrongestAthlete(nickname: string) {
  return nickname === 'KGI' || nickname === 'kgi'
}

export function StrengthGreeting({ nickname }: { nickname: string }) {
  if (!isStrongestAthlete(nickname)) {
    return <span className="text-sm font-medium text-slate-700">{nickname}</span>
  }

  return (
    <span className="strength-celebration">
      <span className="relative z-10 inline-flex border border-amber-300 bg-amber-50 px-2.5 py-1 text-sm font-semibold text-amber-950 shadow-sm" aria-hidden="true">
        <span className="mr-1 text-amber-800">самый сильный</span>
        {nickname}
      </span>
      <span className="sr-only">самый сильный {nickname}</span>
      <span className="strength-celebration__burst" aria-hidden="true">
        {strengthMarks.map((mark) => (
          <span key={mark} className="strength-celebration__mark">
            {mark}
          </span>
        ))}
      </span>
    </span>
  )
}
