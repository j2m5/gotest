import type { ComponentProps } from 'react'

export function Divider({ className, ...props }: ComponentProps<'hr'>) {
  return <hr {...props} className={['ui-divider', className].filter(Boolean).join(' ')} />
}
