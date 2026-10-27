import { apiRequest } from '../../shared/api/client'

export function login(email: string, password: string, rememberMe = false) {
  return apiRequest('post', '/api/auth/login', { body: { email, password, remember_me: rememberMe } })
}

export function register(email: string, login: string, password: string) {
  return apiRequest('post', '/api/auth/register', { body: { email, login, password } })
}

export function logout() {
  return apiRequest('post', '/api/auth/logout')
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

export function resetPassword(token: string, password: string) {
  return apiRequest('post', '/api/auth/reset-password', { body: { token, password } })
}
