export interface User {
    id: number
    email: string
    login: string
    character_id: number | null
}

export interface Character {
    id: number
    name: string
    faction_id: number
    class_id: number
    role: string
    level: number
    experience: number
    avatar: string | null
    hp: number
    mp: number
    hp_current: number
    mp_current: number
}

export interface CatalogEntry {
    id: number
    code: string
    name: string
    description: string | null
    icon: string | null
}
export interface CharacterClass extends CatalogEntry { faction_id: number }
export interface CharacterCatalog { factions: CatalogEntry[]; classes: CharacterClass[] }
export interface CharacterRoster { characters: Character[]; active_character_id: number | null }
export interface CreateCharacterInput { name: string; faction_id: number; class_id: number }

export interface AuthResponse {
    user: User
}

export interface ErrorResponse {
    error: string
    errors?: string[]
}

export interface ValidationErrorResponse {
    errors: string[]
}

export interface MessageResponse {
    message: string
}

export interface RegisterResponse extends MessageResponse {
    verification_sent: boolean
}
