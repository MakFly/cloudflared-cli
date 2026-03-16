// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
  site: 'https://cloudflared.pulseview.app',
  integrations: [
    starlight({
      title: 'cloudflared-project',
      description: 'Production-grade CLI wrapper for Cloudflare Tunnel management',
      customCss: [
        './src/styles/custom.css',
        '@fontsource/jetbrains-mono/400.css',
        '@fontsource/jetbrains-mono/500.css',
        '@fontsource/jetbrains-mono/700.css',
      ],
      pagefind: true,
      expressiveCode: {
        themes: ['github-dark'],
        styleOverrides: {
          borderColor: 'rgba(255, 255, 255, 0.07)',
          borderRadius: '10px',
          codeBackground: '#0c0e16',
          codeFontFamily: "'JetBrains Mono', 'SF Mono', monospace",
          codeFontSize: '0.85rem',
          frames: {
            editorBackground: '#0c0e16',
            terminalBackground: '#0c0e16',
            terminalTitlebarBackground: 'rgba(255, 255, 255, 0.03)',
            terminalTitlebarBorderBottomColor: 'rgba(255, 255, 255, 0.06)',
          },
        },
      },
      components: {
        Header: './src/components/Header.astro',
        Footer: './src/components/Footer.astro',
        ThemeSelect: './src/components/ThemeSelect.astro',
      },
      social: [
        { icon: 'github', label: 'GitHub', href: 'https://github.com/kev/cloudflared-cli' },
      ],
      head: [
        {
          tag: 'style',
          content: `
            :root, :root[data-theme="light"], :root[data-theme="dark"] {
              color-scheme: dark;
            }
            a[aria-current="true"], a[aria-current="page"] {
              color: #ffffff !important;
              background-color: rgba(255, 255, 255, 0.07) !important;
            }
            starlight-theme-select { display: none !important; }
          `,
        },
        {
          tag: 'link',
          attrs: {
            rel: 'preconnect',
            href: 'https://fonts.googleapis.com',
          },
        },
        {
          tag: 'link',
          attrs: {
            rel: 'preconnect',
            href: 'https://fonts.gstatic.com',
            crossorigin: '',
          },
        },
        {
          tag: 'link',
          attrs: {
            rel: 'stylesheet',
            href: 'https://fonts.googleapis.com/css2?family=Instrument+Sans:wght@400;500;600;700&display=swap',
          },
        },
        {
          tag: 'meta',
          attrs: {
            name: 'keywords',
            content: 'cloudflare tunnel, cloudflared, cli, devops, tunnel management, go',
          },
        },
        {
          tag: 'meta',
          attrs: { property: 'og:image', content: '/og-image.png' },
        },
      ],
      sidebar: [
        {
          label: 'Getting Started',
          items: [
            { label: 'Introduction', slug: '' },
            { label: 'Installation', slug: 'guides/installation' },
            { label: 'Quick Start', slug: 'guides/quickstart' },
          ],
        },
        {
          label: 'Guides',
          items: [
            { label: 'Multi-Environment', slug: 'guides/multi-environment' },
          ],
        },
        {
          label: 'Reference',
          items: [
            { label: 'Commands', slug: 'reference/commands' },
            { label: 'Configuration', slug: 'reference/configuration' },
          ],
        },
      ],
      editLink: {
        baseUrl: 'https://github.com/kev/cloudflared-cli/edit/main/docs/',
      },
    }),
  ],
});
