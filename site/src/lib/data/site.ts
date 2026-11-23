export const site = {
  name: 'hanamura miyu',
  role: 'Fullstack developer',
  url: 'https://miyu.pw',
  noindex: false,
  description:
    'Portfolio of hanamura miyu: websites, Minecraft plugins and mods, bots for Telegram and Discord.',
  contacts: {
    telegram: {
      label: 'Telegram',
      handle: '@hanamuramiyu',
      href: 'https://t.me/hanamuramiyu'
    },
    discord: {
      label: 'Discord',
      handle: 'hanamuramiyu'
    },
    github: {
      label: 'GitHub',
      handle: 'hanamuramiyu',
      href: 'https://github.com/hanamuramiyu'
    }
  },
  hours: {
    days: 'Mon–Fri',
    open: '10:00',
    close: '20:00',
    timeZone: 'Asia/Tokyo',
    zoneLabel: 'UTC+9'
  },
  year: 2026
} as const;