import { createContext, type ReactNode } from 'react'

// The app can supply a footer slot without coupling the UI shell to a domain.
export const LayoutFooterContext = createContext<ReactNode>(null)
