/** @type {import('tailwindcss').Config} */
module.exports = {
  important: true,
  content: [
    './pages/**/*.{js,ts,jsx,tsx}',
    './components/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        white: '#ffffff',
        black: '#000000',
        transparent: 'transparent',
        current: 'currentColor',
        gray: {
          100: '#f3f4f6',
          300: '#d1d5db',
        },
        yellow: {
          50: '#fefce8',
          500: '#eab308',
          700: '#a16207',
        },
        blue: {
          500: '#3b82f6',
        },
      },
    },
  },
  plugins: [],
  presets: [require('@navikt/ds-tailwind')],
}
