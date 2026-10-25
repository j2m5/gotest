import axios from 'axios'
import type { AuthResponse, MessageResponse, RegisterResponse, Character, CharacterCatalog, CharacterRoster, CreateCharacterInput } from '../types'

export { authError } from './errors'

export async function login(email: string, password: string): Promise<AuthResponse> {
    const response = await axios.post<AuthResponse>('/api/auth/login', { email, password })

    return response.data
}

export async function register(email: string, login: string, password: string) {
    const response = await axios.post<RegisterResponse>('/api/auth/register', { email, login, password })

    return response.data
}

export async function logout() {
    const response = await axios.post<MessageResponse>('/api/auth/logout')

    return response.data
}

export async function currentUser(signal?: AbortSignal): Promise<AuthResponse> {
    const response = await axios.get<AuthResponse>('/api/auth/me', { signal })
    return response.data
}

export async function verifyEmail(token: string): Promise<MessageResponse> {
    const response = await axios.post<MessageResponse>('/api/auth/verify', { token })
    return response.data
}

export async function resendVerification(email: string, password: string): Promise<MessageResponse> {
    const response = await axios.post<MessageResponse>('/api/auth/resend-verification', { email, password })
    return response.data
}

export async function listCharacters(): Promise<CharacterRoster> {
    return (await axios.get<CharacterRoster>('/api/characters')).data
}
export async function characterCatalog(): Promise<CharacterCatalog> {
    return (await axios.get<CharacterCatalog>('/api/characters/catalog')).data
}
export async function createCharacter(input: CreateCharacterInput): Promise<{ character: Character }> {
    return (await axios.post<{ character: Character }>('/api/characters', input)).data
}
export async function setActiveCharacter(character_id: number | null): Promise<{ character: Character | null }> {
    return (await axios.put<{ character: Character | null }>('/api/characters/active', { character_id })).data
}

