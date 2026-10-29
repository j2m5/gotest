import type { ComponentProps } from 'react'
import { Link } from 'react-router-dom'

type Variant = 'primary' | 'secondary' | 'danger' | 'ghost'
type Appearance = { variant?: Variant; className?: string }

function buttonClass({ variant = 'primary', className }: Appearance) {
  return ['ui-button', `ui-button--${variant}`, className].filter(Boolean).join(' ')
}

type ButtonProps = ComponentProps<'button'> & Appearance & { pending?: boolean }

export function Button({ variant, className, type = 'button', pending = false, disabled, ...props }: ButtonProps) {
  return <button {...props} type={type} className={buttonClass({ variant, className })}
    disabled={disabled || pending} aria-busy={pending || props['aria-busy']} />
}

export function ButtonLink({ variant, className, ...props }: ComponentProps<typeof Link> & Appearance) {
  return <Link {...props} className={buttonClass({ variant, className })} />
}
