import { http } from '../../shared/api/client'
import type { MessageResponse } from '../../shared/api/types'
import type { AuthResponse } from '../../entities/account/types'

export interface RegisterResponse extends MessageResponse {
  verification_sent: boolean
}

export async function login(email: string, password: string, rememberMe = false): Promise<AuthResponse> {
  return (await http.post<AuthResponse>('/auth/login', { email, password, remember_me: rememberMe })).data
}

export async function register(email: string, login: string, password: string): Promise<RegisterResponse> {
  return (await http.post<RegisterResponse>('/auth/register', { email, login, password })).data
}

export async function logout(): Promise<MessageResponse> {
  return (await http.post<MessageResponse>('/auth/logout')).data
}

export async function currentUser(signal?: AbortSignal): Promise<AuthResponse> {
  return (await http.get<AuthResponse>('/auth/me', { signal })).data
}

export async function verifyEmail(token: string): Promise<MessageResponse> {
  return (await http.post<MessageResponse>('/auth/verify', { token })).data
}

export async function resendVerification(email: string, password: string): Promise<MessageResponse> {
  return (await http.post<MessageResponse>('/auth/resend-verification', { email, password })).data
}

export async function forgotPassword(email: string): Promise<MessageResponse> {
  return (await http.post<MessageResponse>('/auth/forgot-password', { email })).data
}

export async function resetPassword(token: string, password: string): Promise<MessageResponse> {
  return (await http.post<MessageResponse>('/auth/reset-password', { token, password })).data
}
