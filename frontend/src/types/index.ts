export interface User {
    id: number
    email: string
    login: string
}

export interface AuthResponse {
    user: User
}

export interface ErrorResponse {
    error: string
}

export interface ValidationErrorResponse {
    errors: string[]
}