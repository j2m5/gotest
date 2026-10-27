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

test('API generation works outside the repository without database configuration', t => {
  const f = fixture(t)
  mkdirSync(path.join(f.root, 'node_modules/openapi-typescript/bin'), { recursive: true })
  f.write('node_modules/openapi-typescript/bin/cli.js', 'console.log(JSON.stringify({args:process.argv.slice(2),cwd:process.cwd()}))')
  const result = f.run('api:generate')
  assert.equal(result.status, 0, result.stderr)
  const probe = JSON.parse(result.stdout.split('\n').find(line => line.startsWith('{')))
  assert.deepEqual(probe.args, ['api/openapi.yaml', '-o', 'frontend/src/shared/api/schema.d.ts'])
  assert.equal(probe.cwd, f.root)
})

test('API and full checks stop at stale types without rewriting or starting downstream tools', t => {
  const f = fixture(t)
  mkdirSync(path.join(f.root, 'node_modules/openapi-typescript/bin'), { recursive: true })
  f.write('node_modules/openapi-typescript/bin/cli.js', `
    if (!process.argv.includes('--check')) throw new Error('must check, not regenerate');
    console.error('Generated types are not up-to-date!'); process.exitCode = 1;
  `)
  f.write('browser', '')
  for (const task of ['api:check', 'check']) {
    const result = f.run(task, { PATH: path.join(f.root, 'missing-bin'), TEST_DATABASE_URL: 'postgres://test', CHARACTER_BROWSER_BINARY: path.join(f.root, 'browser') })
    failure(result, /Generated types are not up-to-date/)
    assert.doesNotMatch(result.stdout + result.stderr, /Go formatting|Go tests|frontend:|could not start/)
  }
})

test('dev API forwards app flags and preserves the repository working directory', t => {
  const f = fixture(t)
  f.write('go-probe.go', `package main
import("encoding/json";"fmt";"os")
func main(){ wd,_:=os.Getwd(); result,_:=json.Marshal(map[string]any{"args":os.Args[1:],"cwd":wd}); fmt.Printf("PROBE:%s\\n",result) }
`)
  const binary = path.join(f.root, process.platform === 'win32' ? 'go.exe' : 'go')
  const build = spawnSync('go', ['build', '-o', binary, path.join(f.root, 'go-probe.go')], { env: environment(), encoding: 'utf8', windowsHide: true })
  assert.equal(build.status, 0, build.stderr)
  const configFile = path.join(f.root, 'configuration with spaces.env')
  const result = f.run('dev:api', { PATH: f.root + path.delimiter + environment().PATH }, '-env-file', configFile)
  assert.equal(result.status, 0, result.stderr)
  const probe = JSON.parse(result.stdout.split('\n').find(line => line.startsWith('PROBE:')).slice(6))
  assert.deepEqual(probe.args, ['run', './cmd/app', '-env-file', configFile])
  assert.equal(probe.cwd, f.root)
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
