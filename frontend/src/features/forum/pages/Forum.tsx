import { Alert, Button, Checkbox, Input, Textarea } from '../../../shared/ui/kit'
import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import { forumAPI } from '../api'
import type { ForumCategoryList, ForumPost, ForumTopic, ForumTopicList, ForumTopicPage, ModerateInput } from '../api'
import type { APIError } from '../../../shared/api/errors'
import { canModerate, canSanction, formatTime, messageLength, roleLabels, validReason } from '../../social/model/policy'
import { useSocialSession } from '../../social/model/useSocialSession'
import AccountNickname from '../../social/ui/AccountNickname'
import SocialLayout from '../../social/ui/SocialLayout'

type Mode = 'categories' | 'category' | 'topic'
type Edit = { target: 'topic' | 'post'; id: number; version: number; body: string; title: string }
const validBody = (text: string) => messageLength(text) >= 1 && messageLength(text) <= 10000 && !text.includes('\0')
const validTitle = (text: string) => messageLength(text) >= 3 && messageLength(text) <= 120 && !text.includes('\0')
export default function Forum({ mode }: { mode: Mode }) {
  const { user, roster } = useAccount()
  const { categoryId, topicId } = useParams()
  return <SocialLayout><ForumContent key={`${user!.id}:${roster?.active_character_id ?? 'none'}:${mode}:${categoryId}:${topicId}`} mode={mode} id={Number(mode === 'category' ? categoryId : topicId)} /></SocialLayout>
}

function ForumContent({ mode, id }: { mode: Mode; id: number }) {
  const { user, roster, pending } = useAccount()
  const { accountId, characterId, run } = useSocialSession()
  const character = roster?.characters.find(item => item.id === characterId)
  const navigate = useNavigate()
  const [categories, setCategories] = useState<ForumCategoryList | null>(null)
  const [list, setList] = useState<ForumTopicList | null>(null)
  const [page, setPage] = useState<ForumTopicPage | null>(null)
  const [cursor, setCursor] = useState<string | null>(null)
  const [after, setAfter] = useState<number | null>(null)
  const [hidden, setHidden] = useState(false)
  const [revision, setRevision] = useState(0)
  const [error, setError] = useState('')
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [busy, setBusy] = useState(false)
  const [cooldown, setCooldown] = useState(false)
  const [edit, setEdit] = useState<Edit | null>(null)
  const [moderation, setModeration] = useState<Omit<ModerateInput, 'reason'> | null>(null)
  const [reason, setReason] = useState('')
  const staff = canModerate(user!.role)
  const invalidId = mode !== 'categories' && (!Number.isSafeInteger(id) || id <= 0)
  const reload = () => setRevision(value => value + 1)
  useEffect(() => {
    let active = true
    const controller = new AbortController()
    const fail = (failure: APIError) => {
      if (!active) return
      setError(failure.message)
      if (failure.status === 404 || failure.status === 403) { setPage(null); setList(null) }
    }
    if (invalidId) return
    if (mode === 'categories') void run(signal => { signal.addEventListener('abort', () => controller.abort(), { once: true }); return forumAPI.categories(accountId, controller.signal) }, result => { if (active) { setCategories(result); setError('') } }, fail)
    else if (mode === 'category') void run(signal => { signal.addEventListener('abort', () => controller.abort(), { once: true }); return forumAPI.topics(accountId, id, cursor, hidden, controller.signal) }, result => { if (active) { setList(result); setError('') } }, fail)
    else void run(signal => { signal.addEventListener('abort', () => controller.abort(), { once: true }); return forumAPI.topic(accountId, id, after, controller.signal) }, result => { if (active) { setPage(result); setError('') } }, fail)
    return () => { active = false; controller.abort() }
  }, [accountId, id, mode, cursor, hidden, after, revision, run, invalidId])
  useEffect(() => { if (!cooldown) return; const timer = setTimeout(() => setCooldown(false), 5000); return () => clearTimeout(timer) }, [cooldown])
  const ban = list?.forum_ban ?? page?.forum_ban
  const expiry = ban?.expires_at
  useEffect(() => { if (!expiry) return; const timeout = Date.parse(expiry) - Date.now() + 250; if (timeout <= 0) return; const timer = setTimeout(() => setRevision(value => value + 1), timeout); return () => clearTimeout(timer) }, [expiry])
  const failWrite = (failure: APIError) => {
    setBusy(false)
    setError(failure.code === 'forum_conflict' ? 'Публикация уже изменена. Черновик сохранён. Обновите страницу и сравните изменения; повторная отправка не выполняется автоматически.' : failure.message)
    if (failure.status === 429) setCooldown(true)
    if (failure.code === 'forum_banned' || failure.code === 'forum_closed') reload()
  }
  const blocked = pending || busy || !characterId || Boolean(ban) || Boolean(page?.topic.closed_at || page?.topic.hidden_at)
  const editingItem = edit?.target === 'topic' ? page?.topic : page?.posts.find(post => post.id === edit?.id)
  const editNeedsAcknowledgement = Boolean(edit && editingItem && edit.version !== editingItem.version)
  const createAllowed = list && (list.category.create_role === 'user' || user!.role === 'administrator' || (list.category.create_role === 'moderator' && staff))
  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (blocked || cooldown || !characterId || !validBody(body) || (mode === 'category' && (!createAllowed || !validTitle(title)))) return
    setBusy(true); setError('')
    if (mode === 'category') void run(signal => forumAPI.create(accountId, { category_id: id, character_id: characterId, title: title.trim(), body: body.trim() }, signal), result => { setBusy(false); navigate(`/forum/topics/${result.topic.id}`) }, failWrite)
    else void run(signal => forumAPI.reply(accountId, { topic_id: id, character_id: characterId, body: body.trim() }, signal), () => { setBusy(false); setBody(''); setCooldown(true); setAfter(null); reload() }, failWrite)
  }
  const save = (event: FormEvent) => {
    event.preventDefault()
    if (!edit || blocked || !characterId || !editingItem || editNeedsAcknowledgement || !validBody(edit.body) || (edit.target === 'topic' && !validTitle(edit.title))) return
    setBusy(true); setError('')
    const common = { character_id: characterId, body: edit.body.trim(), version: edit.version }
    if (edit.target === 'topic') void run(signal => forumAPI.editTopic(accountId, { ...common, topic_id: edit.id, title: edit.title.trim() }, signal), () => { setBusy(false); setEdit(null); reload() }, failWrite)
    else void run(signal => forumAPI.editPost(accountId, { ...common, post_id: edit.id }, signal), () => { setBusy(false); setEdit(null); reload() }, failWrite)
  }
  const moderate = (event: FormEvent) => {
    event.preventDefault()
    if (!moderation || pending || busy || !validReason(reason)) return
    setBusy(true); setError('')
    void run(signal => forumAPI.moderate(accountId, { ...moderation, reason: reason.trim() }, signal), () => { setBusy(false); setModeration(null); setReason(''); reload() }, failWrite)
  }
  const controls = (item: ForumTopic | ForumPost, target: 'topic' | 'post') => <>
    <div className="social-actions">
      {item.user_id === accountId && !blocked && !item.hidden_at && <Button variant="ghost" onClick={() => { setEdit({ target, id: item.id, version: item.version, body: item.body ?? '', title: 'title' in item ? item.title : '' }); setError('') }}>Редактировать</Button>}
      {staff && (user!.role === 'administrator' || item.author_role === 'user' || item.user_id === accountId) && (target === 'topic' ? [item.hidden_at ? 'restore' : 'hide', page?.topic.closed_at ? 'open' : 'close', page?.topic.pinned_at ? 'unpin' : 'pin'] as const : [item.hidden_at ? 'restore' : 'hide'] as const).map(action => <Button variant="ghost" pending={busy || pending} key={action} disabled={pending || busy} onClick={() => { setModeration({ target, target_id: item.id, action }); setReason('') }}>{({ hide: 'Скрыть', restore: 'Восстановить', close: 'Закрыть', open: 'Открыть', pin: 'Закрепить', unpin: 'Открепить' })[action]}</Button>)}
      {canSanction(user!.role, accountId, item.user_id, item.author_role) && <Link to={`/moderation?user_id=${item.user_id}&scope=forum`} state={{ targetName: item.author_name, targetRole: item.author_role }}>Ограничить {item.author_name}</Link>}
    </div>
    {edit?.id === item.id && edit.target === target && <form className="social-form" onSubmit={save}>
      {edit.version !== item.version && <Alert variant="warning" role="status"><p>Публикация изменилась. Сравните текущий текст выше с вашим черновиком ниже. Подтвердите продолжение, чтобы сохранить свой вариант.</p><Button pending={busy || pending} type="button" disabled={Boolean(blocked)} onClick={() => setEdit({ ...edit, version: item.version })}>Продолжить с моим текстом</Button></Alert>}
      {target === 'topic' && <><label htmlFor="forum-edit-title">Заголовок темы</label><Input id="forum-edit-title" value={edit.title} disabled={busy || pending} onChange={event => setEdit({ ...edit, title: event.target.value })} /></>}
      <label htmlFor="forum-edit-body">Текст изменения</label><Textarea id="forum-edit-body" rows={5} value={edit.body} disabled={busy || pending} onChange={event => setEdit({ ...edit, body: event.target.value })} />
      <small>Текст: 1–10000 символов. Автор останется прежним.</small><div className="social-actions"><Button pending={busy || pending} type="submit" disabled={Boolean(blocked) || edit.version !== item.version || !validBody(edit.body) || (target === 'topic' && !validTitle(edit.title))}>Сохранить изменения</Button><Button variant="secondary" pending={busy} type="button" disabled={busy} onClick={() => setEdit(null)}>Отмена</Button></div>
    </form>}
    {moderation?.target_id === item.id && moderation.target === target && <form className="social-form" onSubmit={moderate}><label htmlFor="forum-reason">Причина действия</label><Textarea id="forum-reason" rows={2} value={reason} onChange={event => setReason(event.target.value)} disabled={busy || pending} /><small>От 3 до 500 символов; действие попадёт в журнал.</small><div className="social-actions"><Button type="submit" pending={busy || pending} disabled={busy || pending || !validReason(reason)}>Применить действие</Button><Button variant="secondary" pending={busy} type="button" disabled={busy} onClick={() => setModeration(null)}>Отмена</Button></div></form>}
  </>
  return <>
    <h1>{mode === 'categories' ? 'Форум' : mode === 'category' ? list?.category.title ?? 'Раздел форума' : page?.topic.title ?? 'Тема форума'}</h1>
    <nav className="social-toolbar" aria-label="Форум"><Link to="/forum">Все разделы</Link>{page && <Link to={`/forum/categories/${page.topic.category_id}`}>К темам раздела</Link>}<Button variant="ghost" pending={busy || pending} disabled={busy || pending} onClick={reload}>Обновить</Button></nav>
    {(error || invalidId) && <Alert variant="error" role="alert">{invalidId ? 'Раздел или тема не найдены.' : error}</Alert>}
    {ban && <Alert variant="warning" role="status"><strong>Форум: публикации и изменения запрещены.</strong> {ban.reason}<br />До {ban.expires_at ? formatTime(ban.expires_at) : 'снятия ограничения'}. Читать форум можно.</Alert>}
    {!character && mode !== 'categories' && <p role="status">Чтобы писать на форуме, <Link to="/characters">выберите персонажа</Link>. Читать можно без персонажа.</p>}
    {mode === 'categories' && <ul className="forum-categories social-records">{categories?.categories.map(category => <li key={category.id}><Link to={`/forum/categories/${category.id}`}>{category.title}</Link><p className="muted">{category.description}</p>{category.create_role !== 'user' && <small>Темы создаёт {roleLabels[category.create_role].toLowerCase()}.</small>}</li>)}</ul>}
    {mode === 'categories' && <p className="chat-directory-link"><Link to="/chat/moderators">Список модераторов</Link></p>}
    {mode === 'category' && <>
      {staff && <label className="forum-hidden-toggle"><Checkbox checked={hidden} onChange={event => { setList(null); setCursor(null); setHidden(event.target.checked) }} /> Показывать скрытые темы</label>}
      <ol className="forum-topics social-records">{list?.topics.map(topic => <li key={topic.id}><Link to={`/forum/topics/${topic.id}`}>{topic.title}</Link><p className="muted">{topic.pinned_at && 'Закреплена · '}{topic.closed_at && 'Закрыта · '}{topic.hidden_at && 'Скрыта · '}Ответов: {topic.reply_count}</p><Header item={topic} /></li>)}</ol>
      {list?.topics.length === 0 && <p className="muted">Тем пока нет.</p>}
      <div className="social-toolbar">{cursor && <Button variant="secondary" onClick={() => { setList(null); setCursor(null) }}>К новым темам</Button>}{list?.next_cursor && <Button variant="secondary" onClick={() => { setList(null); setCursor(list.next_cursor) }}>Следующие темы</Button>}</div>
      {list && !createAllowed && <p className="muted">Создавать темы в этом разделе может только {roleLabels[list.category.create_role].toLowerCase()}.</p>}
    </>}
    {page && <>
      <article className="forum-topic"><Header item={page.topic} /><p className="muted">{page.topic.pinned_at && 'Закреплена. '}{page.topic.closed_at && 'Тема закрыта. '}{page.topic.hidden_at && 'Тема скрыта.'}</p><p className="forum-body">{page.topic.body ?? 'Тема скрыта модератором.'}</p>{page.topic.edited_at && <small className="muted">Изменено {formatTime(page.topic.edited_at)}</small>}{controls(page.topic, 'topic')}</article>
      <h2>Ответы</h2><ol className="forum-posts social-records">{page.posts.map(post => <li key={post.id} data-post-id={post.id}><Header item={post} />{post.hidden_at && <small className="muted">Ответ скрыт</small>}<p className="forum-body">{post.body ?? 'Ответ скрыт модератором.'}</p>{post.edited_at && <small className="muted">Изменено {formatTime(post.edited_at)}</small>}{controls(post, 'post')}</li>)}</ol>
      {page.posts.length === 0 && <p className="muted">Ответов пока нет.</p>}
      <div className="social-toolbar">{after !== null && <Button variant="secondary" onClick={() => { setPage(null); setAfter(null) }}>Первые ответы</Button>}{page.next_after_id !== null && <Button variant="secondary" onClick={() => { setPage(null); setAfter(page.next_after_id) }}>Следующие ответы</Button>}</div>
    </>}
    {((mode === 'categories' && !categories) || (mode === 'category' && !list) || (mode === 'topic' && !page)) && !error && <p role="status">Загрузка форума…</p>}
    {character && ((mode === 'category' && createAllowed) || (mode === 'topic' && page)) && <form className="social-form" onSubmit={submit}>
      <h2>{mode === 'category' ? 'Новая тема' : 'Ответить'}</h2>
      {mode === 'category' && <><label htmlFor="forum-title">Заголовок (3–120 символов)</label><Input id="forum-title" value={title} onChange={event => setTitle(event.target.value)} disabled={Boolean(blocked)} /></>}
      <label htmlFor="forum-body">Текст от {character.name}</label><Textarea id="forum-body" rows={5} value={body} onChange={event => setBody(event.target.value)} disabled={Boolean(blocked)} />
      <div className="social-toolbar"><small>{messageLength(body)} / 10000 · Только текст</small><Button pending={busy || pending} type="submit" disabled={Boolean(blocked) || cooldown || !validBody(body) || (mode === 'category' && !validTitle(title))}>{busy ? 'Подождите…' : cooldown ? 'Пауза 5 секунд' : mode === 'category' ? 'Создать тему' : 'Отправить ответ'}</Button></div>
    </form>}
  </>
}

function Header({ item }: { item: Pick<ForumPost, 'character_id' | 'author_name' | 'author_role' | 'created_at'> }) {
  return <div className="chat-message-header"><AccountNickname role={item.author_role} characterId={item.character_id}>{item.author_name}</AccountNickname>{item.author_role !== 'user' && <span className="social-role">{roleLabels[item.author_role]}</span>}<time dateTime={item.created_at}>{formatTime(item.created_at)}</time></div>
}
