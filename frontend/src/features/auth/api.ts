import { apiRequest } from '../../shared/api/client.ts'

export function login(email: string, password: string, rememberMe = false, signal?: AbortSignal) {
  return apiRequest('post', '/api/auth/login', { body: { email, password, remember_me: rememberMe }, signal })
}

export function register(email: string, login: string, password: string) {
  return apiRequest('post', '/api/auth/register', { body: { email, login, password } })
}

export function logout(accountId: number, signal?: AbortSignal) {
  return apiRequest('post', '/api/auth/logout', { headers: { 'X-Account-ID': accountId }, signal })
}

export function currentUser(signal?: AbortSignal) {
  return apiRequest('get', '/api/auth/me', { signal })
}

export function verifyEmail(token: string) {
  return apiRequest('post', '/api/auth/verify', { body: { token } })
}

export function resendVerification(email: string, password: string) {
  return apiRequest('post', '/api/auth/resend-verification', { body: { email, password } })
}

export function forgotPassword(email: string) {
  return apiRequest('post', '/api/auth/forgot-password', { body: { email } })
}

export function resetPassword(token: string, password: string, signal?: AbortSignal) {
  return apiRequest('post', '/api/auth/reset-password', { body: { token, password }, signal })
}
