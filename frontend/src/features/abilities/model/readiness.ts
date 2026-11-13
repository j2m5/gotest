import type { components } from '../../../shared/api/schema'
import { projectCharacterResources } from '../../../entities/character/regeneration.ts'

export type AbilityAction = {
  id: number; name: string; description: string; icon: string | null
  cost: number; damage: number; heal: number; ready: boolean; reason: string
  label: string; remainingMs: number; cooldownFraction: number
}

function remaining(deadline: string | null, observedAt: number, elapsed: number) {
  if (!deadline) return 0
  const duration = Date.parse(deadline) - observedAt - Math.max(0, elapsed)
  return Number.isFinite(duration) ? Math.max(0, duration) : Infinity
}

// Reusable projection: no encounter, network request or browser wall clock.
export function projectActions(loadout: components['schemas']['AbilityLoadout'], elapsedMs: number, blocked = ''): AbilityAction[] {
  const observed = Date.parse(loadout.observed_at)
  const resources = projectCharacterResources(loadout.character, elapsedMs)
  const global = remaining(loadout.global_ready_at, observed, elapsedMs)
  return loadout.abilities.map(ability => {
    const personal = remaining(ability.ready_at, observed, elapsedMs)
    const reason = blocked ? 'blocked' : resources.hp <= 0 ? 'defeated' : personal > 0 ? 'cooldown'
      : global > 0 ? 'global_cooldown' : resources.mp < ability.mp_cost ? 'insufficient_energy' : 'ready'
    const time = reason === 'cooldown' ? personal : reason === 'global_cooldown' ? global : 0
    const duration = reason === 'cooldown' ? ability.cooldown_ms : loadout.global_cooldown_ms
    const label = blocked || (reason === 'defeated' ? 'Нет здоровья' : reason === 'insufficient_energy' ? 'Мало энергии'
      : time > 0 ? (reason === 'global_cooldown' ? 'GCD · ' : '') + (time / 1000).toFixed(1) + ' с' : 'Готово')
    return { id: ability.id, name: ability.name, description: ability.description, icon: ability.icon,
      cost: ability.mp_cost, damage: ability.damage, heal: ability.heal, ready: reason === 'ready', reason, label,
      remainingMs: time, cooldownFraction: duration > 0 ? Math.min(1, time / duration) : 0 }
  })
}
