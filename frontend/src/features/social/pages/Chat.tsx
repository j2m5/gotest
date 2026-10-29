import { Alert, Button, Textarea } from '../../../shared/ui/kit'
import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, Navigate } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import { socialAPI } from '../api'
import type { ChatMessage, ChatPage } from '../api'
import { SnapshotFeed, connectChat } from '../model/lifecycle'
import type { ConnectionStatus } from '../model/lifecycle'
import { canModerate, canSanction, formatTime, messageLength, roleLabels, validReason } from '../model/policy'
import { useSocialSession } from '../model/useSocialSession'
import SocialLayout from '../ui/SocialLayout'
import AccountNickname from '../ui/AccountNickname'

export default function Chat() {
  const { user, roster } = useAccount()
  const character = roster?.characters.find(item => item.id === roster.active_character_id)
  if (!character) return <Navigate to="/characters" replace />
  return <SocialLayout><ChatContent key={`${user!.id}:${character.id}`} characterId={character.id} name={character.name} /></SocialLayout>
}

function ChatContent({ characterId, name }: { characterId: number; name: string }) {
  const { user, pending: accountPending } = useAccount()
  const { accountId, run, report, current } = useSocialSession()
  const [page, setPage] = useState<ChatPage | null>(null)
  const [cursor, setCursor] = useState<number | null>(null)
  const [revision, setRevision] = useState(0)
  const [retry, setRetry] = useState(0)
  const [status, setStatus] = useState<ConnectionStatus>('connecting')
  const [error, setError] = useState('')
  const [draft, setDraft] = useState('')
  const [busy, setBusy] = useState(false)
  const [cooldown, setCooldown] = useState(false)
  const [deletion, setDeletion] = useState<number | null>(null)
  const [reason, setReason] = useState('')
  useEffect(() => {
    const feed = new SnapshotFeed((before, signal) => socialAPI.messages(accountId, before, signal), result => { if (current()) { setPage(result); setError('') } }, failure => report(failure, result => setError(result.message)))
    feed.page(cursor)
    const connection = connectChat({ accountId, origin: window.location.origin, refresh: feed.refresh, status: setStatus })
    return () => { feed.stop(); connection.stop() }
  }, [accountId, cursor, revision, retry, report, current])
  useEffect(() => {
    if (!cooldown) return
    const timer = setTimeout(() => setCooldown(false), 2000)
    return () => clearTimeout(timer)
  }, [cooldown])
  const banExpiry = page?.chat_ban?.expires_at
  useEffect(() => {
    if (!banExpiry) return
    const delay = Date.parse(banExpiry) - Date.now() + 250
    if (delay <= 0) return
    const timer = setTimeout(() => setRevision(value => value + 1), delay)
    return () => clearTimeout(timer)
  }, [banExpiry])
  const move = (before: number | null) => { setPage(null); setCursor(before); setError(''); setDeletion(null); setReason('') }
  const send = (event: FormEvent) => {
    event.preventDefault()
    if (busy || cooldown || accountPending || page?.chat_ban || messageLength(draft) < 1 || messageLength(draft) > 1000) return
    setBusy(true); setError('')
    void run(signal => socialAPI.send(accountId, characterId, draft.trim(), signal), () => {
      setBusy(false); setDraft(''); setCooldown(true); setCursor(null); setRevision(value => value + 1)
    }, failure => { setBusy(false); setError(failure.message); if (failure.status === 429) setCooldown(true); if (failure.code === 'chat_banned') setRevision(value => value + 1) })
  }
  const remove = (event: FormEvent) => {
    event.preventDefault()
    if (deletion === null || busy || !validReason(reason)) return
    setBusy(true); setError('')
    void run(signal => socialAPI.delete(accountId, deletion, reason.trim(), signal), () => {
      setBusy(false); setDeletion(null); setReason(''); setRevision(value => value + 1)
    }, failure => { setBusy(false); setError(failure.message) })
  }
  const moderator = canModerate(user!.role)
  const ban = page?.chat_ban
  return <>
    <h1>Общий чат</h1>
    <p className="social-connection" role="status">{status === 'connected' ? 'Связь установлена' : status === 'connecting' ? 'Соединяемся…' : 'Восстанавливаем связь…'} {status !== 'connected' && <Button variant="ghost" onClick={() => setRetry(value => value + 1)}>Повторить</Button>}</p>
    {error && <Alert variant="error" role="alert">{error} <Button variant="ghost" onClick={() => setRevision(value => value + 1)}>Обновить</Button></Alert>}
    {ban && <Alert variant="warning" role="status"><strong>Чат: отправка запрещена.</strong> {ban.reason}<br />До {ban.expires_at ? formatTime(ban.expires_at) : 'снятия ограничения'}. Читать сообщения можно.</Alert>}
    <form className="social-form" onSubmit={send}>
      <label htmlFor="chat-body">Сообщение от {name}</label>
      <Textarea id="chat-body" rows={3} value={draft} onChange={event => setDraft(event.target.value)} disabled={busy || accountPending || Boolean(ban)} aria-describedby="chat-length" />
      <div className="social-toolbar"><span id="chat-length" className="muted">{messageLength(draft)} / 1000</span><Button pending={busy || accountPending} type="submit" disabled={busy || accountPending || Boolean(ban) || cooldown || messageLength(draft) < 1 || messageLength(draft) > 1000}>{busy ? 'Подождите…' : cooldown ? 'Пауза 2 секунды' : 'Отправить'}</Button></div>
    </form>
    <div className="social-toolbar"><span>Новые сообщения сверху</span><Button variant="ghost" onClick={() => { move(null); setRevision(value => value + 1) }}>К новым сообщениям</Button></div>
    {!page && <p role="status">Загрузка сообщений…</p>}
    {page?.messages.length === 0 && <p className="muted">Пока нет сообщений.</p>}
    <ol className="chat-messages">
      {page?.messages.map(message => <li key={message.id} className="chat-message">
        <MessageHeader message={message} />
        <p className={message.body === null ? 'chat-body muted' : 'chat-body'}>{message.body ?? 'Сообщение удалено модератором.'}</p>
        {moderator && <div className="social-actions">
          {message.body !== null && (user!.role === 'administrator' || message.author_role === 'user' || message.user_id === accountId) && <Button variant="ghost" pending={busy || accountPending} type="button" disabled={busy || accountPending} onClick={() => { setDeletion(message.id); setReason('') }}>Удалить</Button>}
          {canSanction(user!.role, accountId, message.user_id, message.author_role) && <Link to={`/moderation?user_id=${message.user_id}`} state={{ targetName: message.author_name, targetRole: message.author_role }}>Ограничить {message.author_name}</Link>}
        </div>}
        {deletion === message.id && <form className="social-form" onSubmit={remove}><label htmlFor="delete-reason">Причина удаления</label><Textarea id="delete-reason" rows={2} value={reason} onChange={event => setReason(event.target.value)} /><small>От 3 до 500 символов</small><div className="social-actions"><Button variant="danger" type="submit" pending={busy || accountPending} disabled={busy || accountPending || !validReason(reason)}>Удалить сообщение</Button><Button variant="secondary" pending={busy} type="button" disabled={busy} onClick={() => setDeletion(null)}>Отмена</Button></div></form>}
      </li>)}
    </ol>
    {page?.next_before_id != null && <Button variant="secondary" type="button" onClick={() => move(page.next_before_id)}>Более ранние сообщения</Button>}
    <p className="chat-directory-link"><Link to="/chat/moderators">Список модераторов</Link></p>
  </>
}

function MessageHeader({ message }: { message: ChatMessage }) {
  return <div className="chat-message-header"><AccountNickname role={message.author_role} characterId={message.character_id}>{message.author_name}</AccountNickname>{message.author_role !== 'user' && <span className="social-role">{roleLabels[message.author_role]}</span>}<time dateTime={message.created_at}>{formatTime(message.created_at)}</time></div>
}
