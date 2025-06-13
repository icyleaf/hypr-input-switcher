import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const config: Config = {
  title: 'Hypr Input Switcher',
  tagline: 'Smart input method switcher for Hyprland',
  favicon: 'img/favicon.ico',

  url: 'https://icyleaf.github.io',
  baseUrl: '/hypr-input-switcher/',

  organizationName: 'icyleaf',
  projectName: 'hypr-input-switcher',

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl: 'https://github.com/icyleaf/hypr-input-switcher/tree/main/docs-site/',
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: 'img/hypr-input-switcher-social-card.png',
    navbar: {
      title: 'Hypr Input Switcher',
      logo: {
        alt: 'Hypr Input Switcher Logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'tutorialSidebar',
          position: 'left',
          label: 'Docs',
        },
        {
          href: 'https://github.com/icyleaf/hypr-input-switcher',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {
              label: 'Getting Started',
              to: '/docs/intro',
            },
            {
              label: 'Installation',
              to: '/docs/installation',
            },
            {
              label: 'Configuration',
              to: '/docs/configuration',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'GitHub Issues',
              href: 'https://github.com/icyleaf/hypr-input-switcher/issues',
            },
            {
              label: 'GitHub Discussions',
              href: 'https://github.com/icyleaf/hypr-input-switcher/discussions',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/icyleaf/hypr-input-switcher',
            },
            {
              label: 'Releases',
              href: 'https://github.com/icyleaf/hypr-input-switcher/releases',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} icyleaf. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'yaml', 'go'],
    },
    colorMode: {
      defaultMode: 'dark',
      disableSwitch: true,
      respectPrefersColorScheme: true,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
