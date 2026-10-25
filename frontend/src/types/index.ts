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
    errors?: string[]
}

export interface ValidationErrorResponse {
    errors: string[]
}

export interface MessageResponse {
    message: string
}

export interface RegisterResponse extends MessageResponse {
    verification_sent: boolean
}
