import { memo, useLayoutEffect, useRef, useState } from 'react'
import type { BattleEvent } from '../model/types'
import { Button } from '../../../shared/ui/kit'
import { formatAmount } from '../../wallet/format'

const timeFormat = new Intl.DateTimeFormat('ru-RU', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
function eventText(event: BattleEvent, name: string) {
  switch (event.kind) {
    case 'started': return `Начался бой: ${name}.`
    case 'npc_ability': return `${name}: ${event.ability_name}. Вы получили ${event.damage} урона.`
    case 'player_ability': return `Вы: ${event.ability_name}.${event.damage ? ` Урон: ${event.damage}.` : ''}${event.healing ? ` Восстановлено здоровья: ${event.healing}.` : ''}`
    case 'won': return `Победа! Опыт: +${formatAmount(event.experience)}, кредиты: +${formatAmount(event.credits)}.`
    case 'retreated': return 'Вы отступили. Противник восстановил здоровье.'
    case 'defeated': return 'Поражение. Здоровье продолжает восстанавливаться.'
    case 'expired': return 'Время боя истекло. Противник снова доступен.'
  }
}
export default memo(function BattleLog({ events, npcName, clearedAt }: { events: BattleEvent[]; npcName: string; clearedAt: string | null }) {
  const container = useRef<HTMLDivElement>(null)
  const [following, setFollowing] = useState(true)
  useLayoutEffect(() => {
    if (following && container.current) container.current.scrollTop = container.current.scrollHeight
  }, [events, following])
  return <section className="battle-log" aria-labelledby="battle-log-title">
    <div className="combat-section-heading"><h2 id="battle-log-title">Журнал боя</h2>
      {!following && <Button variant="ghost" onClick={() => setFollowing(true)}>К новым событиям</Button>}</div>
    {clearedAt && <p className="muted" role="status" data-log-cleared>Срок хранения подробного журнала истёк. Итог боя сохранён.</p>}
    <div className="battle-log-scroll" ref={container} role="log" aria-live="polite" aria-relevant="additions" tabIndex={0}
      onScroll={() => { const node = container.current; if (node) setFollowing(node.scrollHeight - node.clientHeight - node.scrollTop < 32) }}>
      {events.length === 0 ? <p className="muted">Событий пока нет.</p> : <ol>{events.map(event => <li key={event.sequence} data-event-kind={event.kind} data-event-sequence={event.sequence}>
        <time dateTime={event.occurred_at}>{timeFormat.format(new Date(event.occurred_at))}</time>
        <span>{eventText(event, npcName)}</span>
      </li>)}</ol>}
    </div>
  </section>
})
