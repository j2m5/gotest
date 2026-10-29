import assert from 'node:assert/strict'

// Shared assertions against rendered controls, not component names or CSS source.
export async function assertUIControls(evaluate, session, label, expectedFieldBackground) {
  const styles = await evaluate(session, `(() => {
    const fields = [...document.querySelectorAll('input:not([type=checkbox]):not([type=hidden]), select, textarea')];
    return {
      links: [...document.querySelectorAll('a')].map(el => ({ text: el.textContent, decoration: getComputedStyle(el).textDecorationLine })),
      fields: fields.map(el => ({ id: el.id, background: getComputedStyle(el).backgroundColor })),
      buttons: [...document.querySelectorAll('button')].map(el => ({ text: el.textContent, disabled: el.disabled, cursor: getComputedStyle(el).cursor, decoration: getComputedStyle(el).textDecorationLine })),
      dividers: [...document.querySelectorAll('hr, .character-card, .catalog-preview, .resource-panel, .experience-row, .section-menu, .chat-message, .social-records > li, .chat-directory-link, .forum-topic, .site-footer')].map(el => {
        const css = getComputedStyle(el);
        return [css.borderTopWidth, css.borderTopStyle, css.borderTopColor];
      }),
    };
  })()`)
  for (const link of styles.links) assert.equal(link.decoration, 'none', `${label}: link has no underline (${link.text.trim()})`)
  const background = expectedFieldBackground ?? styles.fields[0]?.background
  for (const field of styles.fields) {
    assert.equal(field.background, background, `${label}: all input types share the same background (${field.id})`)
    const channels = field.background.match(/\d+/g)?.slice(0, 3).map(Number)
    assert(channels?.length === 3 && channels.every(value => value < 120), `${label}: fields use the dark palette (${field.id})`)
  }
  for (const button of styles.buttons) {
    assert.equal(button.decoration, 'none', `${label}: action text has no underline (${button.text.trim()})`)
    assert(button.disabled ? ['not-allowed', 'progress'].includes(button.cursor) : button.cursor === 'pointer', `${label}: button cursor communicates availability (${button.text.trim()})`)
  }
  for (const divider of styles.dividers) {
    assert.deepEqual(divider.slice(0, 2), ['1px', 'solid'], `${label}: dividers are thin solid lines`)
    assert.deepEqual(divider, styles.dividers[0], `${label}: dividers share one color and weight`)
  }
  return background
}
