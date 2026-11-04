import { Link } from 'react-router-dom'
import { Button } from '../../../shared/ui/kit'
import { currencyNames, formatAmount } from '../format'
import { useWallet } from '../model/context'
import '../wallet.css'

export default function WalletSummary() {
  const { data, error, reload } = useWallet()
  return <div className="wallet-strip" aria-label="Баланс кошелька">
    {error ? <><p role="alert">{error}</p><Button variant="ghost" onClick={reload}>Повторить загрузку баланса</Button></>
      : !data ? <p className="muted" role="status">Загрузка баланса…</p>
        : <Link to="/wallet" aria-label="Кошелёк: кредиты и монеты картеля">{(['credits', 'cartel_coins'] as const).map(currency => <span key={currency}>
          <span className="muted">{currencyNames[currency]}:</span> <strong data-currency={currency} data-amount={data[currency]}>{formatAmount(data[currency])}</strong>
        </span>)}</Link>}
  </div>
}
