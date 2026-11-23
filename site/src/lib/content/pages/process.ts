export interface ProcessStage {
  number: number;
  name: string;
  description: string;
}

export interface ProcessRevisions {
  intro: string;
  included: readonly string[];
  excluded: readonly string[];
  note: string;
}

export interface ProcessPage {
  intro: string;
  stages: readonly ProcessStage[];
  designVsMarkup: readonly string[];
  revisions: ProcessRevisions;
  rounds: readonly string[];
  payments: readonly string[];
  cancellation: readonly string[];
  guarantee: string;
  closing: string;
}

export const processPage: ProcessPage = {
  intro:
    'Every project goes through the same steps. Below is what happens at each one and what I need from you.',
  stages: [
    {
      number: 1,
      name: 'Brief',
      description:
        'We talk in text: what you need, who it is for, what examples you have. I ask questions, you answer. By the end of this stage there is a list of pages and features and a clear scope of work.'
    },
    {
      number: 2,
      name: 'Structure',
      description:
        'We decide what goes on the page and in what order: first the heading, then the work, then the contacts - or the other way round. This is a list of blocks, without styling. It is the easiest place to notice that something is missing or extra.'
    },
    {
      number: 3,
      name: 'Design',
      description:
        'I build the first screen and send a screenshot. It shows how the site will look: colours, fonts, sizes, spacing between blocks. All visual changes are best made here - three rounds at this stage are included in the price.'
    },
    {
      number: 4,
      name: 'Markup',
      description:
        'I build the remaining pages. The site already opens in the browser: the menu works, the links work, the pages switch. I put it on a temporary link on my server and send screenshots. The text can still be temporary, and forms do not send yet.'
    },
    {
      number: 5,
      name: 'Logic',
      description:
        'If you need forms, a database, an admin panel or payments, I build them here. Before this stage the site looks like it works, but no data is saved.'
    },
    {
      number: 6,
      name: 'Content and checks',
      description:
        'I add the real texts and images. I check the phone, tablet, computer, loading speed and all links.'
    },
    {
      number: 7,
      name: 'Handover',
      description:
        'I hand over the code. After that the site lives on your domain and server: that is your side, along with hosting costs. I can explain how to launch it and where things go. Server, network and infrastructure setup is not something I do.'
    }
  ],
  designVsMarkup: [
    'I do not work in Figma. Instead of a mockup, you get screenshots of what is already built in code: first the first screen, then the other pages.',
    'Design is the look. Markup is when the site already opens in the browser and you can move around it: the menu works, the links go where they should, the pages switch. If you need forms, a database or payments, logic comes after markup.',
    'From the markup stage you get a link to the test stand and look at the site yourself, not only in pictures.'
  ],
  revisions: {
    intro:
      'A revision is a change inside something already approved. A new block, a new page or a new feature does not count as a revision.',
    included: [
      'change the text;',
      'change the colour, font or size within the approved look;',
      'swap blocks around;',
      'replace an image with another one;',
      'fix spacing and alignment;',
      'rewrite a heading or a description.'
    ],
    excluded: [
      'a new page;',
      'a new block that was not in the structure;',
      'a new feature: a form, payment, an account area;',
      'a change of concept after the look is approved;',
      'changes after a stage is accepted and closed.'
    ],
    note:
      'If you are not sure what your request counts as, ask before I start. I will tell you straight: it is a revision or new work.'
  },
  rounds: [
    'A round is one message with the full list of notes. Not "fix this here" and then "and this here too" - all of it at once. It is easier for you, and it is clearer to me what to change.',
    'Three rounds at the design stage and two at the markup stage are included in the price.',
    'A round beyond that starts at €6, and a rush round starts at €10. I give the exact price before I start: it depends on the size of the change and the time.',
    'If you think the rounds will not be enough, say so in advance - we will discuss it.'
  ],
  payments: [
    '40% up front books your slot in my schedule. The remaining 60% is paid before I hand over the source files.',
    'If the project is large, the payment is split in three: 40% at the start, 30% in the middle after the checks and approvals, 30% before handover.',
    'While the work is in progress, the site is on my server. You open the link and look at the site as a normal visitor. The code, server access and settings stay with me until handover.'
  ],
  cancellation: [
    'Cancel before work starts - full refund.',
    'Cancel after work starts - refund minus the finished stage and the hours actually spent.',
    'Cancel because of me - full refund.'
  ],
  guarantee:
    'If something is broken because of me - a bug, a mismatch with the approved look, a form that does not send - I fix it for free within 30 days after handover.',
  closing:
    'If something here is unclear, ask before we start. It is cheaper than fixing it later.'
};
