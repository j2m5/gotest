import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useLocation, useSearchParams } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import { socialAPI } from '../api'
import type { ActionPage, BanInput, BanPage, UserBan } from '../api'
import { SnapshotFeed, connectChat } from '../model/lifecycle'
import { canModerate, canSanction, formatTime, scopeLabels, validReason } from '../model/policy'
import type { Role } from '../model/policy'
import { useSocialSession } from '../model/useSocialSession'
import SocialLayout from '../ui/SocialLayout'

const durations = [1, 3, 6, 12, 24, 72, 168]
const actionLabels = { ban: 'Ограничение', revoke: 'Снятие ограничения', delete_message: 'Удаление сообщения', role_change: 'Смена роли' }
const positiveId = (text: string) => /^\d+$/.test(text) && Number.isSafeInteger(Number(text)) && Number(text) > 0 ? Number(text) : undefined
const activeBan = (ban: UserBan) => !ban.revoked_at && (!ban.expires_at || Date.parse(ban.expires_at) > Date.now())

export default function Moderation() {
  const { user, roster } = useAccount()
  return <SocialLayout>{user && canModerate(user.role)
    ? <ModerationContent key={`${user.id}:${roster?.active_character_id ?? 'none'}`} />
    : <><h1>Модерация</h1><p role="alert">Этот раздел доступен модераторам.</p><Link to="/chat">Вернуться в чат</Link></>}</SocialLayout>
}

function ModerationContent() {
  const { user, pending: accountPending } = useAccount()
  const { accountId, run, report, current } = useSocialSession()
  const [params] = useSearchParams()
  const location = useLocation()
  const initialId = positiveId(params.get('user_id') ?? '')
  const context = location.state as { targetName?: string; targetRole?: Role } | null
  const [target, setTarget] = useState(String(initialId ?? ''))
  const [filter, setFilter] = useState<number | undefined>(initialId)
  const [scope, setScope] = useState<BanInput['scope']>('chat')
  const [duration, setDuration] = useState('1')
  const [reason, setReason] = useState('')
  const [revokeId, setRevokeId] = useState<number | null>(null)
  const [revokeReason, setRevokeReason] = useState('')
  const [bans, setBans] = useState<BanPage | null>(null)
  const [actions, setActions] = useState<ActionPage | null>(null)
  const [banCursor, setBanCursor] = useState<number | null>(null)
  const [actionCursor, setActionCursor] = useState<number | null>(null)
  const [revision, setRevision] = useState(0)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [busy, setBusy] = useState(false)
  const targetId = positiveId(target)
  const targetRole = targetId === initialId ? context?.targetRole : undefined
  const allowed = targetId !== undefined && canSanction(user!.role, accountId, targetId, targetRole)
  useEffect(() => {
    const feed = new SnapshotFeed((_cursor, signal) => Promise.all([
      socialAPI.bans(accountId, banCursor, filter, signal), socialAPI.actions(accountId, actionCursor, signal),
    ]), ([banPage, actionPage]) => { if (current()) { setBans(banPage); setActions(actionPage); setError('') } }, failure => report(failure, result => setError(result.message)))
    feed.refresh()
    const connection = connectChat({ accountId, origin: window.location.origin, refresh: feed.refresh, status: () => {} })
    return () => { feed.stop(); connection.stop() }
  }, [accountId, banCursor, actionCursor, filter, revision, report, current])
  const fail = (failure: { message: string }) => { setError(failure.message); setBusy(false) }
  const issue = (event: FormEvent) => {
    event.preventDefault()
    if (!allowed || !targetId || busy || accountPending || !validReason(reason)) return
    setBusy(true); setError(''); setNotice('')
    void run(signal => socialAPI.ban(accountId, { user_id: targetId, scope, duration_hours: duration === 'permanent' ? null : Number(duration), reason: reason.trim() }, signal), () => {
      setBusy(false); setReason(''); setNotice('Ограничение выдано.'); setBanCursor(null); setActionCursor(null); setRevision(value => value + 1)
    }, fail)
  }
  const revoke = (event: FormEvent) => {
    event.preventDefault()
    if (revokeId === null || busy || accountPending || !validReason(revokeReason)) return
    setBusy(true); setError(''); setNotice('')
    void run(signal => socialAPI.revoke(accountId, revokeId, revokeReason.trim(), signal), () => {
      setBusy(false); setRevokeId(null); setRevokeReason(''); setNotice('Ограничение снято.'); setRevision(value => value + 1)
    }, fail)
  }
  const changeScope = (next: BanInput['scope']) => { setScope(next); if (next !== 'account' && duration === 'permanent') setDuration('1') }
  return <>
    <h1>Модерация</h1>
    {error && <p className="flash-error" role="alert">{error} <button className="link-button" onClick={() => setRevision(value => value + 1)}>Обновить</button></p>}
    {notice && <p className="social-notice" role="status">{notice}</p>}
    <h2>Выдать ограничение</h2>
    <form className="social-form" onSubmit={issue}>
      <label htmlFor="ban-target">ID аккаунта</label><input id="ban-target" inputMode="numeric" value={target} onChange={event => setTarget(event.target.value)} disabled={busy || accountPending} required />
      {targetId === initialId && context?.targetName && <p className="muted">Персонаж: {context.targetName} · аккаунт #{targetId}</p>}
      {targetId === accountId && <p role="status">Свой аккаунт ограничить нельзя.</p>}
      {targetId && !allowed && targetId !== accountId && <p role="status">Недостаточно прав для ограничения этого аккаунта.</p>}
      <label htmlFor="ban-scope">Раздел</label><select id="ban-scope" value={scope} onChange={event => changeScope(event.target.value as BanInput['scope'])} disabled={busy || accountPending}><option value="chat">Чат</option><option value="forum">Форум</option>{user!.role === 'administrator' && <option value="account">Аккаунт</option>}</select>
      <label htmlFor="ban-duration">Срок</label><select id="ban-duration" value={duration} onChange={event => setDuration(event.target.value)} disabled={busy || accountPending}>{durations.map(hours => <option key={hours} value={hours}>{hours} ч.</option>)}{scope === 'account' && user!.role === 'administrator' && <option value="permanent">Бессрочно</option>}</select>
      <label htmlFor="ban-reason">Причина ограничения</label><textarea id="ban-reason" rows={2} value={reason} onChange={event => setReason(event.target.value)} disabled={busy || accountPending} /><small>От 3 до 500 символов. Ограничение действует на весь аккаунт.</small>
      <button disabled={busy || accountPending || !allowed || !validReason(reason)}>{busy ? 'Подождите…' : 'Выдать ограничение'}</button>
    </form>
    <h2>Ограничения</h2>
    <form className="social-toolbar" onSubmit={event => { event.preventDefault(); if (target.trim() && !targetId) return; setFilter(targetId); setBanCursor(null); setBans(null); setRevision(value => value + 1) }}><span>{filter ? `Аккаунт #${filter}` : 'Все аккаунты'}</span><button disabled={Boolean(target.trim()) && !targetId}>Показать для указанного ID</button><button type="button" onClick={() => { setFilter(undefined); setBanCursor(null); setBans(null); setRevision(value => value + 1) }}>Все</button></form>
    {banCursor !== null && <button className="link-button" onClick={() => { setBanCursor(null); setBans(null) }}>К новым ограничениям</button>}
    {!bans && <p role="status">Загрузка ограничений…</p>}
    {bans?.bans.length === 0 && <p className="muted">Ограничений нет.</p>}
    <ol className="social-records">{bans?.bans.map(ban => <li key={ban.id}>
      <strong>#{ban.id} · Аккаунт #{ban.user_id} · {scopeLabels[ban.scope]}</strong>
      <p className="chat-body">{ban.reason}</p><p className="muted">Выдано {formatTime(ban.created_at)} · модератор #{ban.created_by_user_id}<br />Срок: {ban.expires_at ? `до ${formatTime(ban.expires_at)}` : 'бессрочно'} · {ban.revoked_at ? `снято ${formatTime(ban.revoked_at)}` : activeBan(ban) ? 'действует' : 'истекло'}</p>
      {activeBan(ban) && canSanction(user!.role, accountId, ban.user_id, ban.user_id === initialId ? context?.targetRole : undefined) && (user!.role === 'administrator' || ban.scope !== 'account') && <button className="link-button" disabled={busy || accountPending} onClick={() => { setRevokeId(ban.id); setRevokeReason('') }}>Снять ограничение #{ban.id}</button>}
      {revokeId === ban.id && <form className="social-form" onSubmit={revoke}><label htmlFor="revoke-reason">Причина снятия</label><textarea id="revoke-reason" rows={2} value={revokeReason} onChange={event => setRevokeReason(event.target.value)} disabled={busy || accountPending} /><small>От 3 до 500 символов</small><div className="social-actions"><button disabled={busy || accountPending || !validReason(revokeReason)}>Снять ограничение</button><button type="button" disabled={busy} onClick={() => setRevokeId(null)}>Отмена</button></div></form>}
    </li>)}</ol>
    {bans?.next_before_id != null && <button onClick={() => { setBanCursor(bans.next_before_id); setBans(null) }}>Более ранние ограничения</button>}
    <h2>Журнал действий</h2>
    {actionCursor !== null && <button className="link-button" onClick={() => { setActionCursor(null); setActions(null) }}>К новым действиям</button>}
    {!actions && <p role="status">Загрузка журнала…</p>}
    {actions?.actions.length === 0 && <p className="muted">Действий пока нет.</p>}
    <ol className="social-records">{actions?.actions.map(action => <li key={action.id}><strong>{actionLabels[action.action]} · {scopeLabels[action.scope]}</strong><p>Аккаунт #{action.user_id} · модератор {action.actor_user_id === null ? 'не указан' : `#${action.actor_user_id}`}{action.message_id !== null && ` · сообщение #${action.message_id}`}{action.ban_id !== null && ` · ограничение #${action.ban_id}`}</p><p className="chat-body">{action.reason}</p><time className="muted" dateTime={action.created_at}>{formatTime(action.created_at)}</time></li>)}</ol>
    {actions?.next_before_id != null && <button onClick={() => { setActionCursor(actions.next_before_id); setActions(null) }}>Более ранние действия</button>}
  </>
}
