import { Alert, Button } from '../../../shared/ui/kit'
import type { CharacterCatalog } from '../../../entities/character/types'

export default function CatalogFeedback({ catalog, error, retry }: { catalog: CharacterCatalog | null; error: string; retry: () => void }) {
  if (error) return <div className="catalog-feedback"><Alert variant="error" role="alert">{error}</Alert><Button type="button" onClick={retry}>Повторить</Button></div>
  if (!catalog) return <p role="status" className="muted">Загружаем фракции и классы…</p>
  return null
}
