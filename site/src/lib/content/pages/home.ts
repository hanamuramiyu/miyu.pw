export interface AboutDoItem {
  what: string;
  stack: string;
}

export interface AboutFact {
  label: string;
  value: string;
}

export interface AboutToolGroup {
  label: string;
  items: string;
}

export interface HomePageContent {
  heroTagline: string;
  heroIntro: string;
  workIntro: string;
  aboutLines: readonly string[];
  aboutDo: readonly AboutDoItem[];
  aboutTools: readonly AboutToolGroup[];
  aboutFacts: readonly AboutFact[];
  contactIntro: string;
  contactLocation: string;
}

export const homePage: HomePageContent = {
  heroTagline: 'Websites, Minecraft plugins and mods, bots.',
  heroIntro:
    'I build websites from scratch: front end, back end, server. Separately I make Minecraft plugins and mods, and bots for Telegram and Discord.',
  workIntro: 'Selected work. Everything here was built and shipped.',
  aboutLines: [
    "I'm miyu. I build websites, Minecraft plugins and mods. Websites are what I enjoy most, but plugins are where I have the most experience - it is what I have been doing the longest.",
    'I started working publicly in late 2025 and mostly work alone, though I sometimes team up with a couple of people I know. I write sites by hand: no WordPress, no site builders, no templates.',
    'monban is currently one of my most popular public projects, with almost 1,500 installs on Modrinth. I also built a whole site for a Minecraft server hosting provider: Pterodactyl panel integration, billing and more. The hosting closed, but they want to work with me again.',
    'Outside of work I listen to music - J-pop, J-rap, J-rock, sometimes Chinese music, and bbno$. I like iced coffee but reach for green or fruit tea more often. I play visual novels or rhythm games now and then, but most free time goes to anime, films and series.'
  ],
  aboutDo: [
    { what: 'Websites', stack: 'SvelteKit · TypeScript · Go' },
    { what: 'Plugins and mods for Minecraft', stack: 'Gradle · Java' },
    { what: 'Telegram and Discord bots', stack: 'Go' }
  ],
  aboutTools: [
    { label: 'Front end', items: 'SvelteKit, TypeScript' },
    { label: 'Back end', items: 'Go, Python, PostgreSQL, Redis' },
    { label: 'Minecraft', items: 'Java, Gradle' },
    { label: 'Infrastructure', items: 'Docker, nginx, Linux, Git, bun' }
  ],
  aboutFacts: [
    { label: 'time zone', value: 'UTC+9' },
    { label: 'working since', value: 'late 2025' },
    { label: 'languages', value: 'English, Russian' },
    { label: 'availability', value: 'Depends on current workload' }
  ],
  contactIntro:
    'Tell me what you need, your budget and your deadline, and how you would like me to address you. Also say whether you prefer a formal or a casual tone, whether the project can be shown here, and whether the client can be named.',
  contactLocation: 'Based in UTC+9.'
};
