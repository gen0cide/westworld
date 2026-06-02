// Skills/stats tab — driven by /state.self.skills (already served today). A
// first feature beyond raw parity with the old client.

import type { Self } from '../types'

export function StatsPanel({ self }: { self: Self | null }) {
  if (!self) return <div className="stats"><div className="head">stats</div>loading…</div>
  return (
    <div className="stats">
      <div className="head">{`combat ${self.combatLevel} \xB7 qp ${self.questPoints} \xB7 fatigue ${Math.round(self.fatigue / 100)}%`}</div>
      {self.skills.map((s) => (
        <div className="srow" key={s.id}>
          <span>{s.name}</span>
          <span className="lvl">{s.level === s.max ? s.level : `${s.level}/${s.max}`}</span>
        </div>
      ))}
    </div>
  )
}
