export interface User {
  id: number
  email: string
  login: string
  character_id: number | null
}

export interface AuthResponse {
  user: User
}
