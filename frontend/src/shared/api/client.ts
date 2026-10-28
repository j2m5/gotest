import axios from 'axios'
import type { paths } from './schema'

const http = axios.create()

type Method = 'get' | 'post' | 'put' | 'patch' | 'delete'
type APIPath = Extract<keyof paths, `/api/${string}`>
type PathFor<M extends Method> = {
  [P in APIPath]: paths[P][M] extends { responses: unknown } ? P : never
}[APIPath]
type JSONBody<T> = T extends { content: { 'application/json': infer Body } } ? Body : never
type HeaderOptions<Operation> = Operation extends { parameters: { header?: infer Headers } }
  ? { headers?: Headers } : { headers?: never }
type QueryOptions<Operation> = Operation extends { parameters: { query?: infer Query } }
  ? { query?: Query } : { query?: never }
type RequestOptions<Operation> = { signal?: AbortSignal } & HeaderOptions<Operation> & QueryOptions<Operation> & (
  Operation extends { requestBody: infer Body } ? { body: JSONBody<Body> }
    : Operation extends { requestBody?: infer Body } ? { body?: JSONBody<NonNullable<Body>> }
      : { body?: never }
)
type RequestArgs<Operation> = Operation extends { requestBody: unknown }
  ? [options: RequestOptions<Operation>]
  : [options?: RequestOptions<Operation>]
type SuccessStatus<Responses> = {
  [Status in keyof Responses]: `${Status & (string | number)}` extends `2${string}` ? Status : never
}[keyof Responses]
type ResponseBody<Operation> = Operation extends { responses: infer Responses }
  ? JSONBody<Responses[SuccessStatus<Responses>]> : never

// NoInfer keeps the request body from widening the selected method/URL to a
// different operation. The deadline also bounds transports that stall before connecting.
export async function apiRequest<M extends Method, P extends PathFor<M>>(
  method: M,
  path: P,
  ...args: RequestArgs<paths[NoInfer<P>][NoInfer<M>]>
): Promise<ResponseBody<paths[P][M]>> {
  const [options] = args
  const controller = new AbortController()
  let timer: ReturnType<typeof setTimeout> | undefined
  let abort: (() => void) | undefined
  const cancellation = new Promise<never>((_, reject) => {
    abort = () => { reject(new axios.CanceledError()); controller.abort() }
    if (options?.signal?.aborted) { abort(); return }
    options?.signal?.addEventListener('abort', abort, { once: true })
    timer = setTimeout(() => {
      reject(Object.assign(new Error('Request deadline exceeded'), { code: 'ETIMEDOUT' }))
      controller.abort()
    }, 20_000)
  })
  try {
    const response = await Promise.race([
      cancellation,
      http.request<ResponseBody<paths[P][M]>>({
        method, url: path, data: options?.body, params: options?.query, headers: options?.headers, signal: controller.signal,
      }),
    ])
    return response.data
  } finally {
    clearTimeout(timer)
    if (abort) options?.signal?.removeEventListener('abort', abort)
  }
}
