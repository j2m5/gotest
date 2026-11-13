import { Button } from '../../../shared/ui/kit'
import LocalImage from '../../../shared/ui/LocalImage'
import type { AbilityAction } from '../model/readiness'
import '../abilities.css'

// The adapter supplies projected actions; the bar does not know combat targets.
export default function AbilityBar({ actions, onUse }: { actions: AbilityAction[]; onUse: (id: number) => void }) {
  return <div className="ability-bar" role="group" aria-label="Умения персонажа">
    {actions.map(action => <Button key={action.id} className="ability-button" variant="secondary" data-ability-id={action.id}
      aria-disabled={!action.ready} onClick={() => { if (action.ready) onUse(action.id) }} aria-label={`${action.name}, энергия ${action.cost}, ${action.label}`}
      title={`${action.description}\nЭнергия: ${action.cost}. ${action.label}`}>
      <span className="ability-icon"><LocalImage src={action.icon} alt="" width={34} height={34} />
        <span className="ability-cooldown-shade" aria-hidden="true" style={{ height: `${action.cooldownFraction * 100}%` }} /></span>
      <span className="ability-copy"><strong>{action.name}</strong><span>Энергия: {action.cost}</span></span>
      <span className="ability-readiness" aria-hidden="true">{action.label}</span>
    </Button>)}
  </div>
}
