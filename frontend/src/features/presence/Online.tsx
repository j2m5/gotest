import { useEffect, useState } from 'react'
import { useAccount } from '../../entities/account/context'
import { factionArtwork } from '../../entities/character/appearance'
import { Alert, Button } from '../../shared/ui/kit'
import type { components } from '../../shared/api/schema'
import { useCharacterCatalog } from '../character/model/useCharacterCatalog'
import SocialLayout from '../social/ui/SocialLayout'
import AccountNickname from '../social/ui/AccountNickname'
import { SnapshotFeed } from '../social/model/lifecycle'
import { useSocialSession } from '../social/model/useSocialSession'
import { presenceAPI } from './api'

interface OnlinePage { total: number; characters: components['schemas']['PublicCharacterProfile'][]; next_after_id: number | null }
export default function Online() {
  const { user, roster } = useAccount()
  return <SocialLayout><Directory key={`${user!.id}:${roster?.active_character_id ?? 'none'}`} /></SocialLayout>
}
function Directory() {
  const [cursors, setCursors] = useState<(number | null)[]>([null])
  const cursor = cursors[cursors.length - 1]
  return <DirectoryPage key={cursor ?? 'first'} cursor={cursor} back={cursors.length > 1 ? () => setCursors(values => values.slice(0, -1)) : undefined} next={id => setCursors(values => [...values, id])} />
}
function DirectoryPage({ cursor, back, next }: { cursor: number | null; back?: () => void; next: (id: number) => void }) {
  const { accountId, current, report } = useSocialSession()
  const { catalog } = useCharacterCatalog()
  const [page, setPage] = useState<OnlinePage | null>(null)
  const [error, setError] = useState('')
  const [attempt, setAttempt] = useState(0)
  useEffect(() => {
    let observed = -Infinity
    const feed = new SnapshotFeed(
      (after, signal) => presenceAPI.online(accountId, after, signal),
      value => { if (current()) { observed = performance.now(); setPage(value); setError('') } },
      failure => { if (current()) { setPage(null); report(failure, result => setError(result.message)) } },
    )
    const refresh = () => {
      if (document.visibilityState !== 'visible' || !current()) return
      if (performance.now() - observed >= 120000) setPage(null)
      feed.refresh()
    }
    feed.page(cursor)
    const timer = window.setInterval(refresh, 30000)
    window.addEventListener('focus', refresh)
    document.addEventListener('visibilitychange', refresh)
    return () => { feed.stop(); window.clearInterval(timer); window.removeEventListener('focus', refresh); document.removeEventListener('visibilitychange', refresh) }
  }, [accountId, cursor, current, report, attempt])
  return <>
    <h1>Кто онлайн</h1>
    <p className="muted">Персонажи, активные за последние 2 минуты. Всего: {page?.total ?? '—'}</p>
    {error && <Alert variant="error" role="alert">{error}</Alert>}
    {!page && !error && <p role="status">Загрузка списка…</p>}
    {page?.characters.length === 0 && <p className="muted">На этой странице нет персонажей в сети.</p>}
    <ul className="social-records online-directory">{page?.characters.map(character => {
      const faction = catalog?.factions.find(item => item.id === character.faction_id)
      return <li key={character.id}>
        <img className="icon" src={factionArtwork(faction).icon} alt={faction?.name ?? ''} width={25} height={25} />
        <AccountNickname role={character.account_role} characterId={character.id}>{character.name}</AccountNickname>
        <span className="muted">Уровень {character.level}</span>
      </li>
    })}</ul>
    <div className="social-toolbar">
      {back && <Button variant="ghost" onClick={back}>Назад</Button>}
      {page?.next_after_id != null && <Button variant="ghost" onClick={() => next(page.next_after_id!)}>Далее</Button>}
      <Button variant="ghost" onClick={() => { setPage(null); setError(''); setAttempt(value => value + 1) }}>Обновить</Button>
    </div>
  </>
}
