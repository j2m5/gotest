import { useEffect, useState } from 'react'
import type { SubmitEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import { apiError } from '../../../shared/api/errors'
import type { CharacterCatalog } from '../../../entities/character/types'
import type { FieldErrors } from '../../../shared/api/errors'
import FieldFeedback from '../../../shared/ui/FieldFeedback'
import Layout from '../../../shared/ui/Layout'
export default function NewCharacter() {
  const { store, pending } = useAccount()
  const [catalog, setCatalog] = useState<CharacterCatalog | null>(null)
  const [catalogError, setCatalogError] = useState('')
  const [retry, setRetry] = useState(0)
  const [error, setError] = useState('')
  const [fields, setFields] = useState<FieldErrors>({})
  const clearFields = (...names: string[]) => {
    setError('')
    setFields(current => Object.fromEntries(Object.entries(current).filter(([key]) => !names.includes(key))))
  }
  const [name, setName] = useState('')
  const [faction, setFaction] = useState('')
  const [characterClass, setCharacterClass] = useState('')
  const navigate = useNavigate()
  useEffect(() => {
    let active = true
    store.catalog().then(value => { if (active) setCatalog(value) }).catch(error => { if (active) setCatalogError(apiError(error).message) })
    return () => { active = false }
  }, [store, retry])
  const classes = catalog?.classes.filter(item => item.faction_id === Number(faction)) ?? []
  const submit = async (event: SubmitEvent) => {
    event.preventDefault()
    if (pending) return
    setError('')
    setFields({})
    try {
      await store.create({ name, faction_id: Number(faction), class_id: Number(characterClass) })
      navigate('/characters', { replace: true })
    } catch (error) {
      const failure = apiError(error, ['name', 'faction_id', 'class_id'])
      setError(failure.message)
      setFields(failure.fields)
    }
  }
  return <Layout><h1>Создать персонажа</h1>
    {catalogError ? <><p className="flash-error" role="alert">{catalogError}</p><button className="btn-primary" onClick={() => { setCatalogError(''); setRetry(value => value + 1) }}>Повторить</button></> : !catalog ? <p role="status">Загружаем фракции и классы…</p> : !catalog.factions.length ? <p role="alert">Нет доступных фракций.</p> : <form onSubmit={submit} aria-busy={pending}>
      {error && <p className="flash-error" role="alert">{error}</p>}
      <div className="form-item"><label className="block" htmlFor="character-name">Имя</label><input id="character-name" type="text" required disabled={pending} value={name} onChange={event => { setName(event.target.value); clearFields("name") }} aria-invalid={Boolean(fields.name?.length)} aria-describedby={fields.name?.length ? "name-help name-error" : "name-help"} /><FieldFeedback fields={fields} name="name" id="name-error" /><p id="name-help" className="sm-text">3–32 буквы латиницы или кириллицы, пробел, дефис или апостроф. Первая и последняя — буквы.</p></div>
      <div className="form-item"><label className="block" htmlFor="faction">Фракция</label><select id="faction" aria-invalid={Boolean(fields.faction_id?.length)} aria-describedby={fields.faction_id?.length ? "faction-error" : undefined} required disabled={pending} value={faction} onChange={event => { setFaction(event.target.value); setCharacterClass(''); clearFields("faction_id", "class_id") }}><option value="">Выберите фракцию</option>{catalog.factions.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}</select><FieldFeedback fields={fields} name="faction_id" id="faction-error" /></div>
      {catalog.factions.find(item => item.id === Number(faction))?.description && <p>{catalog.factions.find(item => item.id === Number(faction))?.description}</p>}
      <div className="form-item"><label className="block" htmlFor="character-class">Класс</label><select id="character-class" aria-invalid={Boolean(fields.class_id?.length)} aria-describedby={fields.class_id?.length ? "class-error" : undefined} required disabled={pending || !faction} value={characterClass} onChange={event => { setCharacterClass(event.target.value); clearFields("class_id") }}><option value="">Выберите класс</option>{classes.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}</select><FieldFeedback fields={fields} name="class_id" id="class-error" /></div>
      {classes.find(item => item.id === Number(characterClass))?.description && <p>{classes.find(item => item.id === Number(characterClass))?.description}</p>}
      <button className="btn-primary" disabled={pending || !faction || !characterClass} type="submit">{pending ? 'Создаём…' : 'Создать'}</button>
    </form>}
    {!pending && <p><Link to="/characters">К списку персонажей</Link></p>}
  </Layout>
}
