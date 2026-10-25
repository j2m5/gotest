import type { APIError } from '../../shared/api/errors'

export function needsEmailVerification(error: APIError): boolean {
  return error.status === 403 && error.code === 'email_unverified'
}
