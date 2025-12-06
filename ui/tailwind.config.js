/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{svelte,js,ts}"
  ],
  theme: {
    extend: {
      fontFamily: {
        mono: ['JetBrains Mono', 'Menlo', 'Monaco', 'Consolas', 'monospace'],
      },
      colors: {
        // Terminal-inspired dark theme
        terminal: {
          bg: '#1a1b26',
          surface: '#24283b',
          border: '#414868',
          text: '#c0caf5',
          muted: '#565f89',
        },
        // Log level colors
        log: {
          debug: '#7aa2f7',
          info: '#9ece6a',
          warn: '#e0af68',
          error: '#f7768e',
        }
      }
    },
  },
  plugins: [],
}
