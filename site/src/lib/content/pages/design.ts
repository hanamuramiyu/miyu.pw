export interface DesignColor {
  token: string;
  hex: string;
  ink: 'light' | 'dark';
  role: string;
}

export interface DesignTypeStep {
  token: string;
  size: string;
  font: 'display' | 'body';
  purpose: string;
  sample: string;
}

export interface DesignFont {
  name: string;
  usage: string;
  family: 'display' | 'body';
  url: string;
}

export interface DesignSurface {
  name: string;
  token: 'bg' | 'bg-raised';
  raised: boolean;
}

export interface DesignImage {
  src: string;
  width: number;
  height: number;
  alt: string;
  label: string;
}

export interface DesignPage {
  intro: string;
  colour: {
    title: string;
    note: string;
    tokens: readonly DesignColor[];
  };
  type: {
    title: string;
    note: string;
    columns: { font: string; usage: string; sample: string };
    fonts: readonly DesignFont[];
    pair: { display: string; body: string };
    steps: readonly DesignTypeStep[];
  };
  layout: {
    title: string;
    note: string;
    workRow: {
      left: string;
      right: string;
      title: string;
      description: string;
      tools: string;
      period: string;
    };
    sectionsTitle: string;
    sectionsNote: string;
    sections: readonly DesignSurface[];
  };
  character: {
    title: string;
    note: string;
    examples: readonly DesignImage[];
  };
}

export const designPage: DesignPage = {
  intro: 'The rules this site is built on. Everything here is used on every page.',
  colour: {
    title: 'Colour',
    note: 'Eight colours, and no more are needed. Each one is defined once: change it in one place and it changes everywhere.',
    tokens: [
      { token: 'bg', hex: '#141211', ink: 'light', role: 'page background' },
      { token: 'bg-raised', hex: '#1b1816', ink: 'light', role: 'alternating sections' },
      { token: 'text', hex: '#ebe4dc', ink: 'dark', role: 'primary text' },
      { token: 'text-muted', hex: '#a89d94', ink: 'dark', role: 'secondary text' },
      { token: 'line', hex: '#2c2825', ink: 'light', role: 'thin lines' },
      { token: 'line-strong', hex: '#3b3531', ink: 'light', role: 'accent lines' },
      { token: 'accent', hex: '#d5c4b4', ink: 'dark', role: 'links and accents' },
      { token: 'accent-soft', hex: '#e7d9cb', ink: 'dark', role: 'hover' }
    ]
  },
  type: {
    title: 'Type',
    note: 'Two families. Comfortaa holds the name and headings; Manrope carries everything you read.',
    columns: { font: 'Typeface', usage: 'Where it is used', sample: 'Sample' },
    fonts: [
      {
        name: 'Comfortaa',
        usage: 'name, section headings, work titles',
        family: 'display',
        url: 'https://fonts.google.com/specimen/Comfortaa'
      },
      {
        name: 'Manrope',
        usage: 'text, labels, links, navigation',
        family: 'body',
        url: 'https://fonts.google.com/specimen/Manrope'
      }
    ],
    pair: {
      display: 'hanamura miyu',
      body: 'hanamura miyu'
    },
    steps: [
      {
        token: 'xs',
        size: '13px',
        font: 'body',
        purpose: 'facts and tool lists',
        sample: 'SvelteKit · TypeScript · Go · Java · Python · UTC+9'
      },
      {
        token: 'sm',
        size: '15px',
        font: 'body',
        purpose: 'secondary text and navigation',
        sample: 'Selected work. Everything here was built and shipped.'
      },
      {
        token: 'base',
        size: '17px',
        font: 'body',
        purpose: 'body text',
        sample: 'Websites, Minecraft plugins and mods, bots.'
      },
      {
        token: 'lg',
        size: '20px',
        font: 'body',
        purpose: 'lead paragraphs',
        sample: 'I started working publicly in late 2025 and mostly work alone.'
      },
      {
        token: 'xl',
        size: '24px',
        font: 'body',
        purpose: 'emphasised lines',
        sample: 'Fullstack developer'
      },
      {
        token: '2xl',
        size: '30px',
        font: 'display',
        purpose: 'work titles',
        sample: 'monban'
      },
      {
        token: '3xl',
        size: '28–40px',
        font: 'display',
        purpose: 'page headings',
        sample: 'How I work'
      },
      {
        token: 'display',
        size: '40–72px',
        font: 'display',
        purpose: 'the name and the first screen',
        sample: 'hanamura miyu'
      }
    ]
  },
  layout: {
    title: 'Layout',
    note: 'The page takes the full width of the screen, with one gutter at the edges. Sections alternate their background by a couple of shades - enough to separate one from another.',
    workRow: {
      left: 'Title, summary, links',
      right: 'Tools, period',
      title: 'Work title',
      description: 'A one-line description of the work.',
      tools: 'Tools',
      period: 'Nov 2025 - now'
    },
    sectionsTitle: 'Background order',
    sectionsNote: 'The background changes across the full width and only by a couple of shades. It is a rhythm, not blocks.',
    sections: [
      { name: 'first screen', token: 'bg', raised: false },
      { name: 'work', token: 'bg-raised', raised: true },
      { name: 'about', token: 'bg', raised: false },
      { name: 'contact', token: 'bg-raised', raised: true }
    ]
  },
  character: {
    title: 'The character',
    note: 'The artwork is used only as the site icon and the social card.',
    examples: [
      {
        src: '/icon-192.png',
        width: 192,
        height: 192,
        alt: 'Site icon: a fragment of the character amahane yume.',
        label: 'icon-192.png · 192 × 192'
      },
      {
        src: '/og/site.webp',
        width: 1200,
        height: 630,
        alt: 'Social card with the name hanamura miyu on the left and amahane yume on the right.',
        label: 'og/site.webp · 1200 × 630'
      }
    ]
  }
};
