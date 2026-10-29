import { useEffect, useReducer } from 'react'
import type { Character } from './types'
import { characterResources } from './regeneration'

export function useCharacterResources(character: Character | undefined) {
  const [, tick] = useReducer((value: number) => value + 1, 0)
  useEffect(() => {
    if (!character?.regeneration) return
    const timer = window.setInterval(tick, 1000)
    return () => window.clearInterval(timer)
  }, [character])
  // Read the current snapshot every render; no previous character's values are retained.
  return character ? characterResources(character) : { hp: 0, mp: 0 }
}
