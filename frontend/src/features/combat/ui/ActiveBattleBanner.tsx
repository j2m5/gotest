import { useLocation } from 'react-router-dom'
import { Alert, ButtonLink } from '../../../shared/ui/kit'
import { useCombat } from '../model/context'
import '../combat.css'

export default function ActiveBattleBanner() {
  const { snapshot } = useCombat()
  const { pathname } = useLocation()
  const battle = snapshot.current?.battle
  if (!battle || pathname === `/battle/${battle.id}`) return null
  return <Alert variant="warning" className="active-battle-banner">
    <span>Бой продолжается: <strong>{battle.npc.name}</strong></span>
    <ButtonLink to={`/battle/${battle.id}`} variant="secondary">Вернуться в бой</ButtonLink>
  </Alert>
}
