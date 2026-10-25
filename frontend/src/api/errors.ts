export interface AuthError {
  message: string
  status?: number
}

function record(value: unknown): Record<string, unknown> | undefined {
  return typeof value === 'object' && value !== null ? value as Record<string, unknown> : undefined
}

// Axios failures, malformed server responses and unknown thrown values all reach here.
export function authError(error: unknown): AuthError {
  const response = record(record(error)?.response)
  const status = typeof response?.status === 'number' ? response.status : undefined
  const data = record(response?.data)
  const message = typeof data?.error === 'string' && data.error.trim() ? data.error : undefined
  const details = Array.isArray(data?.errors)
    ? data.errors.filter((item): item is string => typeof item === 'string' && Boolean(item.trim()))
    : []

  return {
    status,
    message: [message, details.join('; ')].filter(Boolean).join(': ') || (response
      ? 'Не удалось выполнить запрос. Попробуйте позже: сервер недоступен.'
      : 'Не удалось связаться с сервером. Проверьте соединение и попробуйте ещё раз.'),
  }
}
