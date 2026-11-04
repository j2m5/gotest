import type { ComponentProps } from 'react'
import { Link } from 'react-router-dom'

type BreadcrumbsProps = Omit<ComponentProps<'nav'>, 'children'> & {
  items: readonly { label: string; to?: string }[]
}

export function Breadcrumbs({ items, className, ...props }: BreadcrumbsProps) {
  return <nav aria-label="Хлебные крошки" {...props} className={['ui-breadcrumbs', className].filter(Boolean).join(' ')}>
    <ol role="list">{items.map((item, index) => <li key={item.to ?? item.label}>
      {index > 0 && <span className="ui-breadcrumbs__separator" aria-hidden="true">›</span>}
      {item.to
        ? <Link to={item.to} aria-current={index === items.length - 1 ? 'page' : undefined}>{item.label}</Link>
        : <span aria-current={index === items.length - 1 ? 'page' : undefined}>{item.label}</span>}
    </li>)}</ol>
  </nav>
}
