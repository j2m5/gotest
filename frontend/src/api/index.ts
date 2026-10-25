import axios from 'axios'
import type { AuthResponse, MessageResponse, RegisterResponse } from '../types'

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

