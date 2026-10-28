import type { CharacterCatalog } from '../../../entities/character/types'

export default function CatalogFeedback({ catalog, error, retry }: { catalog: CharacterCatalog | null; error: string; retry: () => void }) {
  if (error) return <div className="catalog-feedback"><p className="flash-error" role="alert">{error}</p><button type="button" className="btn-primary" onClick={retry}>Повторить</button></div>
  if (!catalog) return <p role="status" className="muted">Загружаем фракции и классы…</p>
  return null
}
