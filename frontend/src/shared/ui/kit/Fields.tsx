import type { ComponentProps } from 'react'

function fieldClass(className?: string) {
  return ['ui-field', className].filter(Boolean).join(' ')
}

export function Input({ className, type = 'text', ...props }: ComponentProps<'input'>) {
  return <input {...props} type={type} className={fieldClass(className)} />
}

export function Select({ className, ...props }: ComponentProps<'select'>) {
  return <select {...props} className={fieldClass(className)} />
}

export function Textarea({ className, ...props }: ComponentProps<'textarea'>) {
  return <textarea {...props} className={fieldClass(className)} />
}

export function Checkbox({ className, ...props }: Omit<ComponentProps<'input'>, 'type'>) {
  return <input {...props} type="checkbox" className={['ui-checkbox', className].filter(Boolean).join(' ')} />
}
