import { spawnSync } from 'node:child_process'
import { existsSync, mkdirSync, readdirSync, statSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = fileURLToPath(new URL('../', import.meta.url))
const frontend = path.join(root, 'frontend')
const goPackages = ['./cmd/...', './internal/...', './tests/...']
const commands = ['help', 'setup', 'dev:api', 'dev:web', 'build', 'lint', 'audit', 'audit:go', 'api:generate', 'api:check', 'migrate', 'seed', 'cleanup', 'mail:status', 'db:backup', 'db:restore', 'test', 'test:integration', 'test:race', 'check', 'test:db:up', 'test:db:down']

function run(label, command, args, { cwd = root, capture = false, secrets = [] } = {}) {
  console.log(`> ${label}`)
  const result = spawnSync(command, args, { cwd, env: process.env, windowsHide: true,
    stdio: capture ? ['inherit', 'pipe', 'pipe'] : 'inherit', encoding: 'utf8', maxBuffer: 10 << 20 })
  const redact = text => secrets.filter(Boolean).reduce((value, secret) => value.split(secret).join('[redacted]'), text ?? '')
  if (capture) {
    process.stdout.write(redact(result.stdout))
    process.stderr.write(redact(result.stderr))
  }
  if (result.error) throw new Error(`${label}: could not start or finish tool (${result.error.code})`)
  if (result.status !== 0) throw new Error(`${label} failed (exit ${result.status ?? result.signal})`)
  return result.stdout ?? ''
}

function npm(args, cwd = frontend) {
  // Invoking the Node CLI also works on Windows, without a shell or npm.cmd quoting.
  if (!process.env.npm_execpath) throw new Error('Run this command through npm run (see npm run help)')
  run(`${cwd === root ? 'tooling' : 'frontend'}: npm ${args.join(' ')}`, process.execPath, [process.env.npm_execpath, ...args], { cwd })
}

function apiTypes(check) {
  run(check ? 'Check generated API types' : 'Generate API types', process.execPath, [
    path.join(root, 'node_modules', 'openapi-typescript', 'bin', 'cli.js'),
    'api/openapi.yaml', '-o', 'frontend/src/shared/api/schema.d.ts', ...(check ? ['--check'] : []),
  ])
}

function loadEnvironment(name) {
  const file = path.join(root, name)
  if (!existsSync(file)) return
  try { process.loadEnvFile(file) } // Explicit process variables take precedence.
  catch { throw new Error(`Cannot load ${name}; check its format and permissions`) }
}

function testEnvironment(task) {
  loadEnvironment('.env.test')
  if (['check', 'test:integration', 'test:race'].includes(task)) process.env.REQUIRE_DATABASE_TESTS = '1'
  if (task === 'check') process.env.REQUIRE_BROWSER_TESTS = '1'
  const ci = ['true', '1'].includes(process.env.CI?.trim().toLowerCase())
  if ((ci || process.env.REQUIRE_DATABASE_TESTS === '1') && !process.env.TEST_DATABASE_URL?.trim()) {
    throw new Error('TEST_DATABASE_URL is required; configure .env.test or the environment (see .env.test.example)')
  }
  // The browser has its own smoke test in check; race checks exercise Go + SQL.
  if (task === 'test:race') return
  const browser = process.env.CHARACTER_BROWSER_BINARY?.trim()
  if ((ci || process.env.REQUIRE_BROWSER_TESTS === '1') && !browser) {
    throw new Error('CHARACTER_BROWSER_BINARY is required; set the path to Chrome, Chromium or Edge')
  }
  if (browser) {
    const file = path.resolve(root, browser)
    if (!existsSync(file) || !statSync(file).isFile()) throw new Error('CHARACTER_BROWSER_BINARY must point to an existing browser executable')
    process.env.CHARACTER_BROWSER_BINARY = file
  }
  if (process.env.CHARACTER_SCREENSHOT_DIR?.trim()) {
    process.env.CHARACTER_SCREENSHOT_DIR = path.resolve(root, process.env.CHARACTER_SCREENSHOT_DIR.trim())
  }
}

function build() {
  const directory = path.join(root, '.local', 'bin')
  mkdirSync(directory, { recursive: true })
  run('Build Go binaries', 'go', ['build', '-o', directory + path.sep, './cmd/...'])
  npm(['run', 'build'])
}

function lint() {
  const unformatted = run('Go formatting', 'gofmt', ['-l', 'cmd', 'internal', 'tests'], { capture: true })
  if (unformatted.trim()) throw new Error('Run gofmt on the files listed above')
  run('Go vet', 'go', ['vet', ...goPackages])
  npm(['run', 'lint'])
}

function tests() {
  const toolsTests = readdirSync(path.join(root, 'tests', 'devtools')).filter(name => name.endsWith('.test.mjs'))
    .sort().map(name => path.join(root, 'tests', 'devtools', name))
  run('Development command tests', process.execPath, ['--test', ...toolsTests])
  run('Go tests', 'go', ['test', '-count=1', '-v', './tests/...'])
  npm(['test'])
}

function migrate(action) {
  loadEnvironment('.env')
  const dsn = process.env.DATABASE_URL?.trim()
  if (!dsn) throw new Error('DATABASE_URL is required for application migrations')
  let password
  try {
    const url = new URL(dsn)
    if (!['postgres:', 'postgresql:'].includes(url.protocol) || !url.hostname) throw new Error('Invalid database URL')
    password = decodeURIComponent(url.password)
  }
  catch { throw new Error('DATABASE_URL must be a valid PostgreSQL URL for migrate') }
  // migrate turns -path into a file URL; a Windows drive path is not a valid URL host.
  // Every subprocess already runs from root, so a relative source works on both OSes.
  run(`migrate ${action}`, 'migrate', ['-path', 'migrations', '-database', dsn, '-lock-timeout', '10', action],
    { capture: true, secrets: [dsn, password, encodeURIComponent(password)] })
}

function compose(task) {
  loadEnvironment('.env.test')
  const envFile = existsSync(path.join(root, '.env.test')) ? '.env.test' : '.env.test.example'
  const args = task === 'test:db:up' ? ['up', '-d', '--wait', '--wait-timeout', '60'] : ['down', '--volumes']
  run('Disposable test PostgreSQL', 'docker', ['compose', '--env-file', envFile, '-f', 'compose.test.yml', ...args])
}

function main() {
  const [task = 'help', ...args] = process.argv.slice(2)
  if (!commands.includes(task)) throw new Error(`Unknown command: ${task}; use npm run help`)
  if (task === 'migrate') {
    if (args.length > 1 || (args[0] && !['up', 'version'].includes(args[0]))) throw new Error('migrate accepts only up or version; use the native CLI for an explicit rollback')
  } else if (args.length && !['seed', 'cleanup', 'mail:status', 'db:backup', 'db:restore', 'dev:api'].includes(task)) throw new Error(`${task} does not accept extra arguments`)
  if (['test', 'test:integration', 'test:race', 'check'].includes(task)) testEnvironment(task)
  switch (task) {
    case 'help':
      console.log('Commands: npm run <name>\n' + commands.join('\n'))
      console.log('\ncheck requires test PostgreSQL and Chrome. test permits local integration skips.\nUse .env.test.example; application commands use .env. migrate: up (default) or version.')
      return
    case 'setup':
      run('Go dependencies', 'go', ['mod', 'download'])
      npm(['ci'], root)
      npm(['ci'])
      return
    case 'dev:api': return run('Go API', 'go', ['run', './cmd/app', ...args])
    case 'dev:web': return npm(['run', 'dev'])
    case 'build': return build()
    case 'lint': return lint()
    case 'audit':
      npm(['audit', '--include=dev', '--audit-level=moderate'], root)
      return npm(['audit', '--include=dev', '--audit-level=moderate'])
    case 'audit:go':
      return run('Go vulnerability scan', 'go', ['run', 'golang.org/x/vuln/cmd/govulncheck@v1.9.0', '-test', ...goPackages])
    case 'api:generate': return apiTypes(false)
    case 'api:check':
      apiTypes(true)
      npm(['run', 'test:types'])
      return run('Go HTTP contract tests', 'go', ['test', '-count=1', './tests/routes', '-run', '^TestOpenAPI'])
    case 'migrate': return migrate(args[0] ?? 'up')
    case 'seed': return run('Seed application catalogs', 'go', ['run', './cmd/seed', ...args])
    case 'cleanup': return run('Clean expired credentials and old mail jobs', 'go', ['run', './cmd/cleanup', ...args])
    case 'mail:status': return run('Mail queue status', 'go', ['run', './cmd/mailqueue', ...args])
    case 'db:backup': return run('Back up application database', 'go', ['run', './cmd/dbbackup', 'backup', ...args])
    case 'db:restore': return run('Restore into a separate empty database', 'go', ['run', './cmd/dbbackup', 'restore', ...args])
    case 'check':
      apiTypes(true)
      lint()
      build()
      return tests()
    case 'test':
      if (process.env.CHARACTER_BROWSER_BINARY) npm(['run', 'build'])
      return tests()
    case 'test:integration':
      if (process.env.CHARACTER_BROWSER_BINARY) npm(['run', 'build'])
      return run('PostgreSQL integration tests', 'go', ['test', '-count=1', '-v', './tests/db', './tests/seed'])
    case 'test:race':
      process.env.CGO_ENABLED = '1'
      return run('Go race tests (requires a C compiler)', 'go', ['test', '-race', '-count=1', '-v', '-skip', '^Test(Character|Social)BrowserFlow$', './tests/...'])
    default: return compose(task)
  }
}

try { main() }
catch (error) {
  console.error(error.message)
  process.exitCode = 1
}
