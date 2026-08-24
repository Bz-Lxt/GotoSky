/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        night: '#07090f',
        panel: '#10141e',
        gold: '#f5c451',
        silver: '#c9d4e0',
        bronze: '#c9844a',
        ink: '#e8e4d8',
        mute: '#8b93a7',
        danger: '#e25b5b'
      },
      fontFamily: {
        sans: ['Sora', 'Noto Sans SC', 'system-ui', 'sans-serif'],
        mono: ['IBM Plex Mono', 'ui-monospace', 'monospace']
      }
    }
  },
  plugins: []
}
