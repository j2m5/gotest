import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

interface LayoutProps {
  children: ReactNode
  title?: string
  statusPanel?: ReactNode
  footer?: ReactNode
}

export default function Layout({ children, title, statusPanel, footer }: LayoutProps) {
  return (
    <>
      <a className="skip-link" href="#main-content">К содержимому</a>
      <header className="site-header">
        <Link to="/" aria-label="Star Wars: The Old Republic — главная">
          <img src="/images/swtor_logo.png" alt="Star Wars: The Old Republic" width="160" height="48" />
        </Link>
        {title && <p>{title}</p>}
      </header>
      <main id="main-content" tabIndex={-1}>
        {statusPanel}
        {children}
      </main>
      <footer className="site-footer">
        {footer}
        <p>Star Wars: The Old Republic</p>
      </footer>
    </>
  )
}
