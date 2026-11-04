import { apiRequest } from '../../frontend/src/shared/api/client'

// Compiled, never executed. @ts-expect-error must become an error if the API
// boundary starts accepting invalid calls or loses nullable response fields.
export async function checkAPIContract() {
  const account = await apiRequest('post', '/api/auth/login', {
    body: { email: 'p@example.test', password: 'password', remember_me: true },
  })
  account.user.login.toUpperCase()
  account.user.role satisfies 'user' | 'moderator' | 'administrator'
  await apiRequest('post', '/api/auth/login', { body: { email: 'p@example.test', password: 'password' } })
  await apiRequest('post', '/api/auth/login', { body: { email: 'p@example.test', password: 'password', remember_me: null } })
  await apiRequest('get', '/api/auth/me', { signal: new AbortController().signal })
  await apiRequest('post', '/api/auth/logout')
  await apiRequest('post', '/api/auth/verify', { body: { token: 'token' } })

  const created = await apiRequest('post', '/api/characters', { body: { name: 'Malgus', faction_id: 2, class_id: 8 } })
  created.character.name.toUpperCase()
  const selected = await apiRequest('put', '/api/characters/active', { body: { character_id: null } })
  if (selected.character !== null) selected.character.name.toUpperCase()
  await apiRequest('put', '/api/characters/active', { body: { character_id: 10 } })

  // @ts-expect-error unknown URL
  await apiRequest('get', '/api/typo')
  // @ts-expect-error health endpoints are not frontend API operations
  await apiRequest('get', '/healthz')
  // @ts-expect-error login accepts POST only
  await apiRequest('get', '/api/auth/login')
  // @ts-expect-error method and URL cannot be inferred from a mismatching body
  await apiRequest('post', '/api/auth/login', { body: { name: 'Malgus', faction_id: 2, class_id: 8 } })
  // @ts-expect-error required body
  await apiRequest('post', '/api/auth/login')
  // @ts-expect-error missing password
  await apiRequest('post', '/api/auth/login', { body: { email: 'p@example.test' } })
  // @ts-expect-error boolean, not string
  await apiRequest('post', '/api/auth/login', { body: { email: 'p@example.test', password: 'password', remember_me: 'yes' } })
  // @ts-expect-error reject misspelled request fields
  await apiRequest('post', '/api/auth/login', { body: { email: 'p@example.test', password: 'password', remember: true } })
  // @ts-expect-error GET has no JSON request body
  await apiRequest('get', '/api/auth/me', { body: {} })
  // @ts-expect-error account selection requires an explicit value (null is different from missing)
  await apiRequest('put', '/api/characters/active', { body: {} })
  // @ts-expect-error server-owned fields cannot be sent when creating characters
  await apiRequest('post', '/api/characters', { body: { name: 'Malgus', faction_id: 2, class_id: 8, user_id: 1 } })
  // @ts-expect-error account selection can be null
  account.user.character_id.toFixed()
  // @ts-expect-error clearing selection returns null
  selected.character.name.toUpperCase()
  // @ts-expect-error internal account role is not a public response field
  account.user.role_id.toFixed()
}

export async function checkSocialContract() {
  const profile = await apiRequest('get', '/api/characters/profile', { query: { character_id: 3 }, headers: { 'X-Account-ID': 1 } })
  profile.character.account_role satisfies 'user' | 'moderator' | 'administrator'
  profile.character.avatar?.toUpperCase()
  // @ts-expect-error public profiles never contain account email
  profile.character.email.toUpperCase()
  // @ts-expect-error public profiles never contain private resources
  profile.character.hp.toFixed()
  // @ts-expect-error character ID must be numeric
  await apiRequest('get', '/api/characters/profile', { query: { character_id: '3' } })
  const page = await apiRequest('get', '/api/chat/messages', { query: { before_id: 50 }, headers: { 'X-Account-ID': 1 } })
  page.messages.forEach(message => message.body?.toUpperCase())
  // @ts-expect-error cursor must be numeric
  await apiRequest('get', '/api/chat/messages', { query: { before_id: '50' } })
  // @ts-expect-error reject unsupported query keys
  await apiRequest('get', '/api/chat/messages', { query: { user_id: 1 } })
  // @ts-expect-error ordinary account lookup has no query parameters
  await apiRequest('get', '/api/auth/me', { query: { before_id: 50 } })
  // @ts-expect-error duration must be explicit to prevent accidental permanent ban
  await apiRequest('post', '/api/moderation/bans', { body: { user_id: 1, scope: 'chat', reason: 'Spam' } })
}

export async function checkForumContract() {
  const categories = await apiRequest('get', '/api/forum/categories')
  categories.categories.forEach(category => category.create_role satisfies 'user' | 'moderator' | 'administrator')

  const topics = await apiRequest('get', '/api/forum/topics', {
    query: { category_id: 1, cursor: 'p:20', include_hidden: false },
    headers: { 'X-Account-ID': 1 },
  })
  const summary = topics.topics[0]
  if (summary) {
    summary.closed_at?.toUpperCase()
    // @ts-expect-error summaries do not include the topic body or edit version
    summary.body.toUpperCase()
  }
  topics.next_cursor?.toUpperCase()
  if (topics.forum_ban) topics.forum_ban.scope satisfies 'chat' | 'forum' | 'account'

  const topicPage = await apiRequest('get', '/api/forum/topic', { query: { topic_id: 4, after_id: 10 } })
  topicPage.topic.body?.toUpperCase()
  topicPage.topic.edited_at?.toUpperCase()
  topicPage.posts.forEach(post => post.body?.toUpperCase())
  topicPage.next_after_id?.toFixed()

  const created = await apiRequest('post', '/api/forum/topics', {
    headers: { 'X-Account-ID': 1 },
    body: { category_id: 1, character_id: 3, title: 'A topic', body: 'Opening text' },
  })
  created.topic.version.toFixed()
  const reply = await apiRequest('post', '/api/forum/replies', {
    body: { topic_id: 4, character_id: 3, body: 'A reply' },
  })
  reply.post.body?.toUpperCase()
  const editedTopic = await apiRequest('post', '/api/forum/edit-topic', {
    body: { topic_id: 4, character_id: 3, title: 'Updated topic', body: 'Updated text', version: 2 },
  })
  editedTopic.topic.version.toFixed()
  const editedPost = await apiRequest('post', '/api/forum/edit-reply', {
    body: { post_id: 5, character_id: 3, body: 'Updated reply', version: 2 },
  })
  editedPost.post.version.toFixed()
  const moderated = await apiRequest('post', '/api/forum/moderate', {
    body: { target: 'topic', target_id: 4, action: 'hide', reason: 'Policy violation' },
  })
  moderated.message.toUpperCase()

  // @ts-expect-error category_id is required
  await apiRequest('get', '/api/forum/topics', { query: { cursor: 'p:20' } })
  // @ts-expect-error cursor must be a string
  await apiRequest('get', '/api/forum/topics', { query: { category_id: 1, cursor: 20 } })
  // @ts-expect-error GET query keys are strict
  await apiRequest('get', '/api/forum/topic', { query: { topic_id: 4, unknown: true } })
  // @ts-expect-error forum writes reject unknown request properties
  await apiRequest('post', '/api/forum/topics', { body: { category_id: 1, character_id: 3, title: 'A topic', body: 'Text', user_id: 5 } })
  // @ts-expect-error version is required for optimistic edits
  await apiRequest('post', '/api/forum/edit-reply', { body: { post_id: 5, character_id: 3, body: 'Updated reply' } })
  // @ts-expect-error target and action are constrained enums
  await apiRequest('post', '/api/forum/moderate', { body: { target: 'message', target_id: 4, action: 'remove', reason: 'Policy violation' } })
}

export async function checkAccountHeaders() {
  await apiRequest('get', '/api/characters', { headers: { 'X-Account-ID': 1 } })
  // @ts-expect-error header is a numeric account ID
  await apiRequest('get', '/api/characters', { headers: { 'X-Account-ID': '1' } })
  // @ts-expect-error reject misspelled header names
  await apiRequest('get', '/api/characters', { headers: { 'X-Account-Idd': 1 } })
  // @ts-expect-error login does not support account preconditions
  await apiRequest('post', '/api/auth/login', { headers: { 'X-Account-ID': 1 }, body: { email: 'a', password: 'b' } })
  const { listCharacters, createCharacter, characterCatalog, setActiveCharacter } = await import('../../frontend/src/features/character/api')
  const { logout } = await import('../../frontend/src/features/auth/api')
  await listCharacters(1)
  await createCharacter(1, { name: 'Valid', faction_id: 1, class_id: 1 })
  await characterCatalog(1)
  await setActiveCharacter(1, null)
  await logout(1)
  // @ts-expect-error first-party calls always require explicit account IDs
  await listCharacters()
  // @ts-expect-error logout never guesses an account
  await logout()
}

export async function checkWorldContract() {
  const enemies = await apiRequest('get', '/api/npcs', { query: { character_id: 3, location_id: 2 }, headers: { 'X-Account-ID': 1 } })
  const money = await apiRequest('get', '/api/wallet', { query: { character_id: 3 } })
  money.credits satisfies string
  const ledger = await apiRequest('get', '/api/wallet/history', { query: { character_id: 3, before: '9007199254740993' } })
  ledger.entries.forEach(entry => entry.amount satisfies string)
  // @ts-expect-error integer balances are decimal strings, never JS numbers
  money.cartel_coins.toFixed()
  // @ts-expect-error event keys are private server details
  ledger.entries[0].operation_key
  // @ts-expect-error no public currency mutation
  await apiRequest('post', '/api/wallet', { body: { amount: 100 } })
  enemies.npcs.forEach(enemy => enemy.kind satisfies 'humanoid' | 'droid' | 'beast')
  // @ts-expect-error current location is required
  await apiRequest('get', '/api/npcs', { query: { character_id: 3 } })
  // @ts-expect-error instances are shared, not owned by a character
  enemies.npcs[0].character_id.toFixed()
  const navigation = await apiRequest('get', '/api/world/navigation', {
    query: { character_id: 3 }, headers: { 'X-Account-ID': 1 },
  })
  navigation.position.location.planet_id.toFixed()
  navigation.destinations.forEach(destination => destination.can_travel satisfies boolean)
  const moved = await apiRequest('post', '/api/world/travel', {
    headers: { 'X-Account-ID': 1 }, body: { character_id: 3, planet_id: 5, from_location_id: 2 },
  })
  moved.position.location.id.toFixed()
  moved.returned_home satisfies boolean
  const locations = await apiRequest('get', '/api/world/locations', { query: { character_id: 3, planet_id: 5 } })
  locations.locations.forEach(location => location.recommended_level.toFixed())
  await apiRequest('post', '/api/world/move', { body: { character_id: 3, location_id: 7, from_location_id: 2 } })
  // @ts-expect-error browsing is scoped to the requested planet
  await apiRequest('get', '/api/world/locations', { query: { character_id: 3 } })
  // @ts-expect-error a local move must carry its starting location
  await apiRequest('post', '/api/world/move', { body: { character_id: 3, location_id: 7 } })
  // @ts-expect-error navigation is scoped to the expected selected character
  await apiRequest('get', '/api/world/navigation')
  // @ts-expect-error options without the required query are also rejected
  await apiRequest('get', '/api/world/navigation', { headers: { 'X-Account-ID': 1 } })
  // @ts-expect-error movement must carry its starting position
  await apiRequest('post', '/api/world/travel', { body: { character_id: 3, planet_id: 5 } })
  // @ts-expect-error the server derives faction access, never the client
  await apiRequest('post', '/api/world/travel', { body: { character_id: 3, planet_id: 5, from_location_id: 2, faction_id: 1 } })
}
