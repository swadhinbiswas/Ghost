// @ts-check
import { defineConfig } from 'astro/config';

// https://astro.build/config
// Served at the site root (http://localhost:4321/). For a GitHub Pages
// *project* site, deploy behind a custom domain or the user/org pages repo so
// the root path is preserved.
export default defineConfig({
  site: 'https://ghost-cli.dev',
  trailingSlash: 'ignore',
});
