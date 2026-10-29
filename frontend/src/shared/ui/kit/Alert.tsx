import type { ComponentProps } from 'react'

type AlertProps = ComponentProps<'div'> & { variant?: 'info' | 'success' | 'error' | 'warning' }

export function Alert({ variant = 'info', className, role, ...props }: AlertProps) {
  return <div {...props} role={role ?? (variant === 'error' ? 'alert' : 'status')}
    className={['ui-alert', `ui-alert--${variant}`, className].filter(Boolean).join(' ')} />
}
