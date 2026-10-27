import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'

const entityBoundary = {
  regex: '(^|/)(app|features)(/|$)',
  message: 'Entities cannot import application composition or feature implementations.',
}

const featureBoundary = (sibling) => ({
  regex: `(^|/)(app|features/${sibling})(/|$)|^(?:\\.\\./)+${sibling}(/|$)`,
  message: 'Features use entities/shared; sibling features are composed by app.',
})

export default defineConfig([
  globalIgnores(['dist', 'src/shared/api/schema.d.ts']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      globals: globals.browser,
    },
  },
  {
    files: ['src/shared/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-imports': ['error', { patterns: [{
        regex: '(^|/)(app|features|entities)(/|$)',
        message: 'Shared code cannot depend on application or domain modules.',
      }] }],
    },
  },
  {
    files: ['src/entities/**/*.{ts,tsx}'],
    rules: { 'no-restricted-imports': ['error', { patterns: [entityBoundary] }] },
  },
  {
    files: ['src/entities/account/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-imports': ['error', { patterns: [entityBoundary, {
        regex: '(^|/)character(/|$)',
        allowTypeImports: true,
        message: 'The account aggregate uses character contracts through type-only imports.',
      }] }],
    },
  },
  {
    files: ['src/entities/character/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-imports': ['error', { patterns: [entityBoundary, {
        regex: '(^|/)account(/|$)',
        message: 'Character contracts cannot depend on the account aggregate.',
      }] }],
    },
  },
  {
    files: ['src/features/auth/**/*.{ts,tsx}'],
    rules: { 'no-restricted-imports': ['error', { patterns: [featureBoundary('character')] }] },
  },
  {
    files: ['src/features/character/**/*.{ts,tsx}'],
    rules: { 'no-restricted-imports': ['error', { patterns: [featureBoundary('auth')] }] },
  },
])
