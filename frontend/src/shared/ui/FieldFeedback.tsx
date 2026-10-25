import type { FieldErrors } from '../api/errors'

export default function FieldFeedback({ fields, name, id }: { fields: FieldErrors; name: string; id: string }) {
  const messages = fields[name]
  if (!messages?.length) return null
  return <div id={id} className="flash-error" role="alert">
    {messages.map((message, index) => <p key={index}>{message}</p>)}
  </div>
}
