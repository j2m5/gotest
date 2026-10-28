import type { Character } from './types'

function ResourceBar({ label, value, max, kind }: { label: string; value: number; max: number; kind: 'health' | 'energy' }) {
  const maximum = Math.max(0, max)
  const current = Math.max(0, Math.min(value, maximum))
  const percentage = maximum === 0 ? 0 : Math.round(current / maximum * 100)
  return <div className={`resource-bar resource-${kind}`} role="meter" aria-label={label}
    aria-valuemin={0} aria-valuemax={maximum || 1} aria-valuenow={current} aria-valuetext={`${current} из ${maximum}`}>
    <span className="resource-fill" style={{ width: `${percentage}%` }} />
    <span className="resource-caption" aria-hidden="true"><span>{label}: {current} / {maximum}</span><span>{percentage}%</span></span>
  </div>
}

export default function ResourceBars({ character }: { character: Character }) {
  return <div className="resource-panel" aria-label="Ресурсы персонажа">
    <ResourceBar label="Здоровье" value={character.hp_current} max={character.hp} kind="health" />
    <ResourceBar label="Энергия" value={character.mp_current} max={character.mp} kind="energy" />
  </div>
}
