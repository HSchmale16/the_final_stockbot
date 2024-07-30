/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
	  "base.css",
    "internal/app/html_templates/**/*.hbs",
    "internal/app/html_templates/*.hbs"
  ],
  theme: {
    extend: {},
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/aspect-ratio'),
    require('@tailwindcss/typography'),
    // require('tailwindcss-children'),
  ],
}

