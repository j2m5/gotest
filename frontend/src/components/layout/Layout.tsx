import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

interface LayoutProps {
  children: ReactNode
  title?: string
}

export default function Layout({ children, title }: LayoutProps) {
  return (
    <>
      <header>
        <Link to="/">{title || 'Logo'}</Link>
      </header>
      <main>
        {children}
      </main>
    </>
  )
}