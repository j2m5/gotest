import { useNavigate } from 'react-router-dom'
import { Alert, Button } from '../../../shared/ui/kit'
import { useCombat } from '../model/context'

export default function CombatFeedback() {
  const { session, snapshot } = useCombat()
  const navigate = useNavigate()
  const retry = async () => {
    const result = await session.retry()
    if (result) navigate(`/battle/${result.battle.id}`)
  }
  if (!snapshot.actionError && !snapshot.unconfirmed) return null
  return <Alert variant={snapshot.unconfirmed ? 'warning' : 'error'}>
    {snapshot.unconfirmed ? <>Ответ на действие не получен. Можно безопасно повторить тот же запрос.
      <div className="combat-actions"><Button pending={snapshot.pending !== null} onClick={() => void retry()}>Повторить действие</Button></div></>
      : snapshot.actionError}
  </Alert>
}
