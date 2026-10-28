// API image fields are optional; an unusable value falls back to a bundled image.
export function localImagePath(value: string | null | undefined): string | undefined {
  if (!value || !/^\/?images\/(?:[a-zA-Z0-9_-]+\/)*[a-zA-Z0-9_-]+\.(?:png|jpe?g|gif|webp)$/.test(value)) return undefined
  return value.startsWith('/') ? value : `/${value}`
}
