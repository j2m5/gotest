import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { copyFileSync, existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const runner = fileURLToPath(new URL('../../scripts/dev.mjs', import.meta.url))
function environment(extra = {}) {
  const excluded = new Set(['ci', 'database_url', 'test_database_url', 'character_browser_binary', 'require_database_tests', 'require_browser_tests', 'npm_execpath', 'path'])
  return { ...Object.fromEntries(Object.entries(process.env).filter(([key]) => !excluded.has(key.toLowerCase()))), PATH: process.env.PATH ?? process.env.Path, ...extra }
}
function fixture(t) {
  const root = mkdtempSync(path.join(tmpdir(), 'gotest commands '))
  assert.equal(path.dirname(path.resolve(root)), path.resolve(tmpdir()))
  t.after(() => rmSync(root, { recursive: true, force: true }))
  mkdirSync(path.join(root, 'scripts'))
  mkdirSync(path.join(root, 'frontend'))
  copyFileSync(runner, path.join(root, 'scripts/dev.mjs'))
  return {
    root,
    write(name, contents) { writeFileSync(path.join(root, name), contents) },
    run(task, env = {}, ...args) {
      return spawnSync(process.execPath, [path.join(root, 'scripts/dev.mjs'), task, ...args], {
        cwd: tmpdir(), env: environment(env), encoding: 'utf8', windowsHide: true, timeout: 30000,
      })
    },
  }
}
function failure(result, text) {
  assert.notEqual(result.status, 0, result.stdout)
  assert.match(result.stderr, text)
}

test('full check requires an explicit test database and never falls back to app .env', t => {
  const f = fixture(t)
  f.write('.env', 'DATABASE_URL=postgres://app-private-value\nTEST_DATABASE_URL=postgres://also-private\n')
  const result = f.run('check', { DATABASE_URL: 'postgres://inherited-app-private-value' })
  failure(result, /TEST_DATABASE_URL/)
  assert.doesNotMatch(result.stdout + result.stderr, /private-value|also-private/)
})

test('test database file is loaded, while explicit environment settings take priority', t => {
  const f = fixture(t)
  f.write('.env.test', 'TEST_DATABASE_URL=postgres://test-file\n')
  failure(f.run('check'), /CHARACTER_BROWSER_BINARY/)
  f.write('.env.test', 'TEST_DATABASE_URL=\n')
  failure(f.run('check', { TEST_DATABASE_URL: 'postgres://explicit-test' }), /CHARACTER_BROWSER_BINARY/)
})

test('full check rejects an invalid browser before running any checks', t => {
  const f = fixture(t)
  failure(f.run('check', { TEST_DATABASE_URL: 'postgres://test', CHARACTER_BROWSER_BINARY: path.join(f.root, 'missing-browser') }), /CHARACTER_BROWSER_BINARY/)
})

test('integration task and direct CI test task both require database configuration', t => {
  const f = fixture(t)
  failure(f.run('test:integration'), /TEST_DATABASE_URL/)
  failure(f.run('test', { CI: 'true' }), /TEST_DATABASE_URL/)
})

test('unsafe migration arguments are rejected without launching migrate', t => {
  const f = fixture(t)
  failure(f.run('migrate', {}, 'down', '-all'), /up.*version/)
})

test('migration requires a PostgreSQL URL before launching the external tool', t => {
  const f = fixture(t)
  failure(f.run('migrate', { DATABASE_URL: 'file:///tmp/not-a-database', PATH: path.join(f.root, 'missing-bin') }, 'version'), /valid PostgreSQL URL/)
})

test('unknown commands fail and help works outside the repository directory', t => {
  const f = fixture(t)
  failure(f.run('does-not-exist'), /Unknown command/)
  const result = f.run('help')
  assert.equal(result.status, 0, result.stderr)
  assert.match(result.stdout, /check/)
})

test('setup stops after a failed first tool instead of invoking the next command', t => {
  const f = fixture(t)
  const marker = path.join(f.root, 'npm-ran')
  const npmCLI = path.join(f.root, 'npm-cli.cjs')
  f.write('npm-cli.cjs', `require('node:fs').writeFileSync(${JSON.stringify(marker)}, 'ran')`)
  const result = f.run('setup', { PATH: path.join(f.root, 'missing-bin'), npm_execpath: npmCLI })
  failure(result, /Go dependencies/)
  assert.equal(existsSync(marker), false)
})

test('migration failures redact credentials from tool output and retain failure status', t => {
  const f = fixture(t)
  const binary = path.join(f.root, process.platform === 'win32' ? 'migrate.exe' : 'migrate')
  f.write('migrate.go', 'package main\nimport("fmt";"os")\nfunc main(){fmt.Println(os.Args);fmt.Fprintln(os.Stderr,"private-password");os.Exit(7)}\n')
  const build = spawnSync('go', ['build', '-o', binary, path.join(f.root, 'migrate.go')], { env: environment(), encoding: 'utf8', windowsHide: true })
  assert.equal(build.status, 0, build.stderr)
  const dsn = 'postgres://user:private-password@localhost/test'
  f.write('.env', `DATABASE_URL=${dsn}\n`)
  const result = f.run('migrate', { PATH: f.root + path.delimiter + environment().PATH }, 'version')
  failure(result, /exit 7/)
  assert.doesNotMatch(result.stdout + result.stderr, /private-password|postgres:\/\//)
})

test('migration source resolves from a repository path with spaces on Windows and Linux', t => {
  const f = fixture(t)
  mkdirSync(path.join(f.root, 'migrations'))
  f.write('migrations/fixture.sql', 'migration-source-loaded')
  // Characterize migrate's file-source URL parsing at the external CLI boundary.
  f.write('migrate.go', `package main
import("fmt";"net/url";"os";"path/filepath")
func main(){
  source:=os.Args[2]
  if _,err:=url.Parse("file://"+source);err!=nil{fmt.Fprintln(os.Stderr,"invalid file URL");os.Exit(2)}
  data,err:=os.ReadFile(filepath.Join(source,"fixture.sql"))
  if err!=nil{fmt.Fprintln(os.Stderr,"source not found");os.Exit(3)}
  fmt.Print(string(data))
}`)
  const binary = path.join(f.root, process.platform === 'win32' ? 'migrate.exe' : 'migrate')
  const build = spawnSync('go', ['build', '-o', binary, path.join(f.root, 'migrate.go')], { env: environment(), encoding: 'utf8', windowsHide: true })
  assert.equal(build.status, 0, build.stderr)
  const result = f.run('migrate', { DATABASE_URL: 'postgres://test@localhost/test', PATH: f.root + path.delimiter + environment().PATH }, 'version')
  assert.equal(result.status, 0, result.stderr)
  assert.match(result.stdout, /migration-source-loaded/)
})
