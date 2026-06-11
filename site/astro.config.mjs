// @ts-check
import { defineConfig } from 'astro/config';

// Local dev (and custom-domain deploys) serve at the root '/'.
// The GitHub Pages *project* site lives under '/Ghost', so the deploy workflow
// sets DEPLOY_TARGET=pages to build with that base. This keeps
// http://localhost:4321/ clean while production assets resolve correctly.
const isPages = process.env.DEPLOY_TARGET === 'pages';

export default defineConfig({
  site: 'https://swadhinbiswas.github.io',
  base: isPages ? '/Ghost' : '/',
  trailingSlash: 'ignore',
});
