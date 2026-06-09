import axios from 'axios'
import type { AuthResponse } from '../types'

export async function login(email: string, password: string): Promise<AuthResponse> {
    const response = await axios.post<AuthResponse>('api/auth/login', { email, password })

    return response.data
}

export async function register(email: string, login: string, password: string) {
    const response = await axios.post('api/auth/register', { email, login, password })

    return response.data
}

export async function logout() {
    const response = await axios.post('api/auth/logout')

    return response.data
}

