export type FieldErrors = Record<string, string[]>

export interface APIError {
  message: string
  status?: number
  code?: string
  fields: FieldErrors
}

function record(value: unknown): Record<string, unknown> | undefined {
  return typeof value === 'object' && value !== null && !Array.isArray(value) ? value as Record<string, unknown> : undefined
}
function messages(value: unknown): string[] {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === 'string' && Boolean(item.trim())) : []
}

// Inline forms omit their rendered field details; other callers retain the full message.
export function apiError(error: unknown, inlineFields: readonly string[] = []): APIError {
  const response = record(record(error)?.response)
  const status = typeof response?.status === 'number' ? response.status : undefined
  const data = record(response?.data)
  const code = typeof data?.code === 'string' && data.code.trim() ? data.code : undefined
  const message = typeof data?.error === 'string' && data.error.trim() ? data.error : undefined
  const fields: FieldErrors = {}
  for (const [key, value] of Object.entries(record(data?.fields) ?? {})) {
    if (['__proto__', 'constructor', 'prototype'].includes(key)) continue
    const details = messages(value)
    if (details.length) fields[key] = [...new Set(details)]
  }
  const inline = new Set(inlineFields.flatMap(key => Object.hasOwn(fields, key) ? fields[key] : []))
  const details = [...new Set([...Object.values(fields).flat(), ...messages(data?.errors)])]
    .filter(item => !inline.has(item) && item !== message)
  return {
    status, code, fields,
    message: [message && !inline.has(message) ? message : undefined, details.join('; ')].filter(Boolean).join(': ') || (inline.size ? '' : response
      ? 'Не удалось выполнить запрос. Попробуйте позже: сервер недоступен.'
      : 'Не удалось связаться с сервером. Проверьте соединение и попробуйте ещё раз.'),
  }
}

export function isUnauthenticated(error: unknown): boolean {
  const failure = apiError(error)
  return failure.status === 401 && (failure.code === undefined || failure.code === 'unauthenticated')
}
