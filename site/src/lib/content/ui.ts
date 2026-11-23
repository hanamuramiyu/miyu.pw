export const ui = {
  skipLink: 'Skip to content',
  nav: {
    label: 'Primary',
    work: 'Work',
    about: 'About',
    process: 'Process',
    design: 'Design',
    contact: 'Contact',
    menu: 'open menu',
    close: 'close menu'
  },
  work: {
    title: 'Work',
    all: 'All work',
    empty: 'No published work yet.',
    palette: 'Palette',
    type: 'Type',
    factsLabel: 'Work facts',
    whatLabel: 'What this is',
    kindLabel: 'Type',
    audienceLabel: 'Audience',
    toolsLabel: 'Tools',
    periodLabel: 'Period',
    roleLabel: 'Role',
    clientLabel: 'Client',
    contentsLabel: 'Contents',
    linksLabel: 'Links',
    galleryLabel: 'Pictures'
  },
  about: {
    title: 'About',
    doTitle: 'What I do',
    toolsTitle: 'Tools'
  },
  process: {
    title: 'How I work',
    stagesTitle: 'Stages',
    designVsMarkupTitle: 'Design and markup',
    revisionsTitle: 'What counts as a revision',
    includedTitle: 'What counts',
    excludedTitle: 'What does not count',
    roundsTitle: 'Revision rounds',
    paymentsTitle: 'Payment',
    cancellationTitle: 'Cancellation',
    guaranteeTitle: 'Guarantee'
  },
  design: {
    title: 'Design'
  },
  contact: {
    title: 'Contact',
    nometaBefore: 'First time writing? Read ',
    nometaLink: 'nometa.xyz',
    nometaAfter: '.',
    copy: 'copy',
    copied: 'copied',
    copyFailed: 'copy failed',
    copyAria: 'Copy Discord handle'
  },
  hours: {
    local: 'my time'
  },
  notFound: {
    heading: 'Nothing here.',
    message: 'The page at this link does not exist. It may have moved, or the link is wrong.',
    homeLink: 'back to home'
  },
  error: {
    heading: 'Something broke.',
    message: 'The page did not load. Go back home and try opening it again.',
    homeLink: 'back to home'
  }
} as const;
