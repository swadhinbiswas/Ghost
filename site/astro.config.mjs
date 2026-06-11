// @ts-check
import { defineConfig } from 'astro/config';

// GitHub Pages deploys under '/Ghost', Cloudflare Pages at root '/'.
const deploy = process.env.DEPLOY_TARGET;
const base = deploy === 'pages' ? '/Ghost' : '/';

export default defineConfig({
  site: deploy === 'pages'
    ? 'https://swadhinbiswas.github.io'
    : 'https://ghostcli.pages.dev',
  base,
  trailingSlash: 'ignore',
});
