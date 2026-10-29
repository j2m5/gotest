// Passive timers intentionally do not enter this event bridge.
export function connectPresenceActivity(
  activity: (foregroundReturn?: boolean) => void,
  browser: Pick<Window, 'addEventListener' | 'removeEventListener'>,
  page: Pick<Document, 'visibilityState' | 'addEventListener' | 'removeEventListener'>,
) {
  const input = () => { if (page.visibilityState === 'visible') activity() }
  const foreground = () => { if (page.visibilityState === 'visible') activity(true) }
  browser.addEventListener('focus', foreground)
  page.addEventListener('visibilitychange', foreground)
  page.addEventListener('pointerdown', input, { passive: true })
  page.addEventListener('keydown', input)
  // Scroll events also come from content refreshes and scroll anchoring. The
  // originating wheel/touch/keyboard/pointer input is the activity boundary.
  page.addEventListener('wheel', input, { passive: true, capture: true })
  page.addEventListener('touchmove', input, { passive: true, capture: true })
  return () => {
    browser.removeEventListener('focus', foreground)
    page.removeEventListener('visibilitychange', foreground)
    page.removeEventListener('pointerdown', input)
    page.removeEventListener('keydown', input)
    page.removeEventListener('wheel', input, { capture: true })
    page.removeEventListener('touchmove', input, { capture: true })
  }
}
