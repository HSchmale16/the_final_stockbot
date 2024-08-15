/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
	  "base.css",
    "internal/m/html_templates/**/*.hbs",
    "internal/m/html_templates/*.hbs",
    "internal/votes/html_templates/**/*.hbs",
    "internal/votes/html_templates/*.hbs",
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

