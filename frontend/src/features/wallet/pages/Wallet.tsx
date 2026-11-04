import { Link, Navigate } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import { Alert, Button } from '../../../shared/ui/kit'
import { useCharacterCatalog } from '../../character/model/useCharacterCatalog'
import GameLayout from '../../character/ui/GameLayout'
import { currencyNames, formatAmount } from '../format'
import { useWallet } from '../model/context'
import { useWalletHistory } from '../model/useWalletHistory'
import '../wallet.css'

export default function Wallet() {
  const { roster } = useAccount()
  const { catalog } = useCharacterCatalog()
  const character = roster?.characters.find(item => item.id === roster.active_character_id)
  if (!character) return <Navigate to="/characters" replace />
  return <GameLayout character={character} catalog={catalog}><WalletContent /></GameLayout>
}

function WalletContent() {
  const balance = useWallet()
  const { data, error, moreError, morePending, reload, loadMore } = useWalletHistory(balance.accountId, balance.characterId)
  const refresh = () => { balance.reload(); reload() }
  return <div className="wallet-page">
    <h1>Кошелёк</h1>
    <p className="muted">Кредиты принадлежат текущему персонажу. Монеты картеля общие для всех персонажей аккаунта.</p>
    <div className="wallet-actions"><Button variant="secondary" onClick={refresh} disabled={!data && !error}>Обновить кошелёк</Button><Link to="/game">На главную</Link></div>
    <h2>История операций</h2>
    {error ? <><Alert variant="error">{error}</Alert><Button onClick={refresh}>Повторить</Button></>
      : !data ? <p role="status">Загрузка истории…</p>
        : data.entries.length === 0 ? <p className="muted">Операций пока нет.</p>
          : <ul className="wallet-history" aria-label="История кошелька">{data.entries.map(entry => <li key={entry.id} data-entry-id={entry.id}>
            <div className="wallet-entry-heading"><strong>{entry.description}</strong><span className={`wallet-amount wallet-amount--${entry.amount.startsWith('-') ? 'debit' : 'credit'}`}>{formatAmount(entry.amount, true)}</span></div>
            <p className="wallet-entry-meta"><span>{currencyNames[entry.currency]}</span><time dateTime={entry.created_at}>{new Date(entry.created_at).toLocaleString('ru-RU', { dateStyle: 'short', timeStyle: 'short' })}</time></p>
            <p className="wallet-entry-balance muted">Остаток: {formatAmount(entry.balance_after)}</p>
          </li>)}</ul>}
    {moreError && <Alert variant="error">{moreError}</Alert>}
    {data?.next_cursor && <Button variant="secondary" pending={morePending} disabled={morePending} onClick={() => void loadMore()}>{moreError ? 'Повторить загрузку' : 'Показать ещё'}</Button>}
  </div>
}
