module.exports = {
  content: ["./**/*.html", "./**/*.templ", "./**/*.go"],
  theme: {
    extend: {
      colors: {
        dark: {
          900: '#0a0a0a',
          800: '#121212',
          700: '#1a1a1a',
          600: '#262626',
          500: '#333333',
        },
        primary: {
          DEFAULT: '#f97316',
          hover: '#ea580c',
          light: '#fb923c',
        },
        accent: {
          DEFAULT: '#22c55e',
          hover: '#16a34a',
          light: '#4ade80',
        },
      },
    },
  },
  plugins: [],
}
