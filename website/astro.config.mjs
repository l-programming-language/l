import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import { unified } from '@astrojs/markdown-remark';
import fs from 'node:fs';
import { focusTables } from './scripts/readable-code.mjs';

const sidebar = JSON.parse(fs.readFileSync(new URL('./.generated/sidebar.json', import.meta.url), 'utf8'));
const identity = JSON.parse(fs.readFileSync(new URL('./.generated/version.json', import.meta.url), 'utf8'));
export default defineConfig({
  site: 'https://ldocs.org',
  base: '/',
  trailingSlash: 'always',
  output: 'static',
  compressHTML: false,
  devToolbar: { enabled: false },
  markdown: { processor: unified({ smartypants: false, remarkPlugins: [focusTables] }) },
  integrations: [starlight({
    title: 'L Documentation',
    description: 'Learn L: a human-readable language with applications as values. Source-backed guides, language specification and compiler references.',
    logo: { dark: './public/assets/l/brand/mark-on-dark.svg', light: './public/assets/l/brand/mark-on-light.svg', alt: 'L', replacesTitle: false },
    favicon: '/assets/l/favicons/favicon.svg',
    locales: { root: { label: 'English', lang: 'en' } },
    sidebar,
    customCss: ['./src/styles/l-theme.css'],
    pagefind: true,
    tableOfContents: { minHeadingLevel: 2, maxHeadingLevel: 3 },
    components: { Footer: './src/components/Footer.astro', Head: './src/components/Head.astro', Search: './src/components/Search.astro' },
    head: [
      { tag: 'meta', attrs: { name: 'robots', content: identity.channel === 'production' ? 'index, follow' : 'noindex, nofollow' } },
      { tag: 'meta', attrs: { property: 'og:image', content: 'https://ldocs.org/assets/l/social/og-default.png' } },
      { tag: 'meta', attrs: { name: 'twitter:card', content: 'summary_large_image' } },
      { tag: 'link', attrs: { rel: 'apple-touch-icon', href: '/assets/l/favicons/apple-touch-icon.png' } },
    ],
  })],
});
