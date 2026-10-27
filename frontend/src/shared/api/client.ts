import axios from 'axios'
import type { paths } from './schema'

const http = axios.create()

type Method = 'get' | 'post' | 'put' | 'patch' | 'delete'
type APIPath = Extract<keyof paths, `/api/${string}`>
type PathFor<M extends Method> = {
  [P in APIPath]: paths[P][M] extends { responses: unknown } ? P : never
}[APIPath]
type JSONBody<T> = T extends { content: { 'application/json': infer Body } } ? Body : never
type RequestOptions<Operation> = { signal?: AbortSignal } & (
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
// different operation. Axios remains responsible for cookies, errors and aborts.
export async function apiRequest<M extends Method, P extends PathFor<M>>(
  method: M,
  path: P,
  ...args: RequestArgs<paths[NoInfer<P>][NoInfer<M>]>
): Promise<ResponseBody<paths[P][M]>> {
  const [options] = args
  const response = await http.request<ResponseBody<paths[P][M]>>({
    method, url: path, data: options?.body, signal: options?.signal,
  })
  return response.data
}
