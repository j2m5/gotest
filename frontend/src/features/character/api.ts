import { http } from '../../shared/api/client'
import type { Character, CharacterCatalog, CharacterRoster, CreateCharacterInput } from '../../entities/character/types'

export async function listCharacters(): Promise<CharacterRoster> {
  return (await http.get<CharacterRoster>('/characters')).data
}

export async function characterCatalog(): Promise<CharacterCatalog> {
  return (await http.get<CharacterCatalog>('/characters/catalog')).data
}

export async function createCharacter(input: CreateCharacterInput): Promise<{ character: Character }> {
  return (await http.post<{ character: Character }>('/characters', input)).data
}

export async function setActiveCharacter(character_id: number | null): Promise<{ character: Character | null }> {
  return (await http.put<{ character: Character | null }>('/characters/active', { character_id })).data
}
