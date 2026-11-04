const integerFormat = new Intl.NumberFormat('ru-RU')
export const currencyNames = { credits: 'Кредиты', cartel_coins: 'Монеты картеля' }

export function formatAmount(value: string, signed = false) {
  const amount = BigInt(value)
  return (signed && amount > 0n ? '+' : '') + integerFormat.format(amount)
}
