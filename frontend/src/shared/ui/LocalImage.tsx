import { localImagePath } from './localImagePath'

interface Props {
  src?: string | null
  fallback?: string
  alt: string
  width: number
  height: number
  className?: string
}

export default function LocalImage({ src, fallback, alt, width, height, className }: Props) {
  const placeholder = localImagePath(fallback) ?? '/images/icons/neutral.png'
  return <img src={localImagePath(src) ?? placeholder} alt={alt} width={width} height={height} className={className}
    onError={event => {
      // A missing placeholder must not create a request loop.
      if (event.currentTarget.getAttribute('src') !== placeholder) event.currentTarget.src = placeholder
    }} />
}
