import tsParser from '@typescript-eslint/parser'
import nextConfig from 'eslint-config-next/core-web-vitals'
import storybookPlugin from 'eslint-plugin-storybook'
import { defineConfig } from 'eslint/config'

export default defineConfig([
  ...nextConfig,
  ...storybookPlugin.configs['flat/recommended'],
  {
    files: ['**/*.{js,jsx,mjs,cjs,ts,tsx,mts,cts}'],
    languageOptions: {
      parser: tsParser,
      parserOptions: { jsx: true },
    },
    settings: {
      react: { version: '19' },
    },
  },
])
