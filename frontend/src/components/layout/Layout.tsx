import type { ReactNode } from 'react'

interface LayoutProps {
  children: ReactNode
  title?: string
}

export default function Layout({ children, title }: LayoutProps) {
  return (
    <>
      <header>
        <a href="/">{title || 'Logo'}</a>
      </header>
      <main>
        {children}
      </main>
    </>
  )
}