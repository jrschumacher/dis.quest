/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./components/**/*.{templ,go}",
    "./server/**/*.go",
    "./internal/**/*.go"
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          blue: '#2563eb',
          'blue-dark': '#1d4ed8',
          'blue-light': '#3b82f6',
          background: '#f4f8fb',
          'dev-banner': '#e0edff',
          'dev-text': '#1e40af'
        }
      },
      fontFamily: {
        sans: ['system-ui', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'Helvetica Neue', 'Arial', 'Noto Sans', 'sans-serif', 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol', 'Noto Color Emoji']
      }
    },
  },
  plugins: [],
}