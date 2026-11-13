import type { components } from '../../../shared/api/schema'
import type { APIError } from '../../../shared/api/errors'

export type BattleState = components['schemas']['BattleState']
export type BattleDetail = components['schemas']['BattleDetail']
export type BattleResult = components['schemas']['BattleResult']
export type BattleEvent = components['schemas']['BattleEvent']
export type Loadout = components['schemas']['AbilityLoadout']
export type Character = components['schemas']['Character']
export type Command = { kind: 'start'; locationId: number; npcId: number; key: string }
  | { kind: 'use'; battleId: string; abilityId: number; key: string }
  | { kind: 'retreat'; battleId: string }
export interface CombatAPI {
  current(signal: AbortSignal): Promise<BattleState>
  detail(id: string, after: number, signal: AbortSignal): Promise<BattleDetail>
  abilities(signal: AbortSignal): Promise<Loadout>
  begin(locationId: number, npcId: number, key: string, signal: AbortSignal): Promise<BattleResult>
  use(id: string, abilityId: number, key: string, signal: AbortSignal): Promise<BattleResult>
  retreat(id: string, signal: AbortSignal): Promise<BattleResult>
}
export interface CombatScope { current(): boolean; report(error: APIError): void }
export interface CombatSnapshot {
  current: BattleState | null; character: Character | null
  detail: BattleDetail | null; loadout: Loadout | null; events: BattleEvent[]
  receivedAt: number; loading: boolean; error: string; actionError: string
  pending: Command['kind'] | null; unconfirmed: boolean
}
