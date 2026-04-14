/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,ts,tsx}'],
  theme: {
    extend: {
      colors: {
        neon: 'var(--neon)',
      },
      backdropBlur: {
        md: '12px',
      }
    }
  },
  plugins: []
}
