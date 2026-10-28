import { useState } from 'react'
import type { SubmitEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import { apiError } from '../../../shared/api/errors'
import type { FieldErrors } from '../../../shared/api/errors'
import FieldFeedback from '../../../shared/ui/FieldFeedback'
import Layout from '../../../shared/ui/Layout'
import LocalImage from '../../../shared/ui/LocalImage'
import { useCharacterCatalog } from '../model/useCharacterCatalog'
import CatalogFeedback from '../ui/CatalogFeedback'
export default function NewCharacter() {
  const { store, pending } = useAccount()
  const catalogState = useCharacterCatalog()
  const { catalog } = catalogState
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
  const classes = catalog?.classes.filter(item => item.faction_id === Number(faction)) ?? []
  const selectedFaction = catalog?.factions.find(item => item.id === Number(faction))
  const selectedClass = classes.find(item => item.id === Number(characterClass))
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
  return <Layout><img className="scene-image" src="/images/scenes/register.jpg" width="550" height="120" alt="" /><h1>Создать персонажа</h1>
    <CatalogFeedback {...catalogState} />
    {catalog && (!catalog.factions.length ? <p role="alert">Нет доступных фракций.</p> : <form onSubmit={submit} aria-busy={pending}>
      {error && <p className="flash-error" role="alert">{error}</p>}
      <div className="form-item"><label className="block" htmlFor="character-name">Имя</label><input id="character-name" type="text" required disabled={pending} value={name} onChange={event => { setName(event.target.value); clearFields("name") }} aria-invalid={Boolean(fields.name?.length)} aria-describedby={fields.name?.length ? "name-help name-error" : "name-help"} /><FieldFeedback fields={fields} name="name" id="name-error" /><p id="name-help" className="sm-text">3–32 буквы латиницы или кириллицы, пробел, дефис или апостроф. Первая и последняя — буквы.</p></div>
      <div className="form-item"><label className="block" htmlFor="faction">Фракция</label><select id="faction" aria-invalid={Boolean(fields.faction_id?.length)} aria-describedby={fields.faction_id?.length ? "faction-error" : undefined} required disabled={pending} value={faction} onChange={event => { setFaction(event.target.value); setCharacterClass(''); clearFields("faction_id", "class_id") }}><option value="">Выберите фракцию</option>{catalog.factions.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}</select><FieldFeedback fields={fields} name="faction_id" id="faction-error" /></div>
      {selectedFaction && <div className="catalog-preview faction-preview">
        <LocalImage src={selectedFaction.icon} alt={selectedFaction.name} width={90} height={90} />
        {selectedFaction.description && <p>{selectedFaction.description}</p>}
      </div>}
      <div className="form-item"><label className="block" htmlFor="character-class">Класс</label><select id="character-class" aria-invalid={Boolean(fields.class_id?.length)} aria-describedby={fields.class_id?.length ? "class-error" : undefined} required disabled={pending || !faction} value={characterClass} onChange={event => { setCharacterClass(event.target.value); clearFields("class_id") }}><option value="">Выберите класс</option>{classes.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}</select><FieldFeedback fields={fields} name="class_id" id="class-error" /></div>
      {selectedClass && <div className="catalog-preview class-preview">
        <LocalImage src={selectedClass.icon} alt={selectedClass.name} width={195} height={120} />
        {selectedClass.description && <p>{selectedClass.description}</p>}
      </div>}
      <button className="btn-primary" disabled={pending || !faction || !characterClass} type="submit">{pending ? 'Создаём…' : 'Создать'}</button>
    </form>)}
    {!pending && <p><Link to="/characters">К списку персонажей</Link></p>}
  </Layout>
}
