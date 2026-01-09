import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'AgentTrace',
  tagline: 'Observability Platform for AI Coding Agents',
  favicon: 'img/favicon.ico',

  url: 'https://docs.agenttrace.io',
  baseUrl: '/',

  organizationName: 'agenttrace',
  projectName: 'agenttrace',

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
          editUrl: 'https://github.com/agenttrace/agenttrace/tree/main/docs/',
          routeBasePath: '/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: 'img/social-card.png',
    navbar: {
      title: 'AgentTrace',
      logo: {
        alt: 'AgentTrace Logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Docs',
        },
        {
          to: '/api-reference',
          label: 'API Reference',
          position: 'left',
        },
        {
          to: '/sdks',
          label: 'SDKs',
          position: 'left',
        },
        {
          href: 'https://github.com/agenttrace/agenttrace',
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
              to: '/getting-started',
            },
            {
              label: 'API Reference',
              to: '/api-reference',
            },
            {
              label: 'SDKs',
              to: '/sdks',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Discord',
              href: 'https://discord.gg/agenttrace',
            },
            {
              label: 'Twitter',
              href: 'https://twitter.com/agenttrace',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/agenttrace/agenttrace',
            },
            {
              label: 'Status',
              href: 'https://status.agenttrace.io',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} AgentTrace. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'python', 'go', 'typescript', 'json', 'yaml'],
    },
    algolia: {
      appId: 'YOUR_APP_ID',
      apiKey: 'YOUR_SEARCH_API_KEY',
      indexName: 'agenttrace',
      contextualSearch: true,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
