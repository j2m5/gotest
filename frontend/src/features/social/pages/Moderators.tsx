import { Alert, Button } from '../../../shared/ui/kit'
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import { socialAPI } from '../api'
import type { ModeratorList } from '../api'
import { roleLabels } from '../model/policy'
import { useSocialSession } from '../model/useSocialSession'
import AccountNickname from '../ui/AccountNickname'
import SocialLayout from '../ui/SocialLayout'

export default function Moderators() {
  const { user, roster } = useAccount()
  return <SocialLayout><Directory key={`${user!.id}:${roster?.active_character_id ?? 'none'}`} /></SocialLayout>
}

function Directory() {
  const { accountId, current, report } = useSocialSession()
  const [list, setList] = useState<ModeratorList | null>(null)
  const [error, setError] = useState('')
  const [revision, setRevision] = useState(0)
  useEffect(() => {
    const controller = new AbortController()
    void socialAPI.moderators(accountId, controller.signal).then(result => {
      if (!controller.signal.aborted && current()) setList(result)
    }, failure => {
      if (!controller.signal.aborted) report(failure, result => setError(result.message))
    })
    return () => controller.abort()
  }, [accountId, current, report, revision])
  return <>
    <h1>Список модераторов</h1>
    {error && <Alert variant="error" role="alert">{error}</Alert>}
    {!list && !error && <p role="status">Загрузка списка…</p>}
    {list?.moderators.length === 0 && <p className="muted">Пока нет модераторов.</p>}
    <ul className="social-records moderator-directory">
      {list?.moderators.map(member => <li key={member.user_id}>
        <div className="chat-message-header"><AccountNickname role={member.role} characterId={member.character_id}>{member.login}</AccountNickname><span className="social-role">{roleLabels[member.role]}</span></div>
        <p className="muted">{member.character_name ? <>Персонаж: <AccountNickname role={member.role} characterId={member.character_id}>{member.character_name}</AccountNickname></> : 'Персонаж не выбран'}</p>
      </li>)}
    </ul>
    <div className="social-toolbar">
      <Link to="/chat">Вернуться в чат</Link>
      <Button variant="ghost" type="button" disabled={!list && !error} onClick={() => { setList(null); setError(''); setRevision(value => value + 1) }}>Обновить список</Button>
    </div>
  </>
}
