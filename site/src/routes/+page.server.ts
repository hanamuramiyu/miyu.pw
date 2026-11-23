import type { PageServerLoad } from './$types';
import { getFeaturedWorks, toWorkListItem } from '#lib/content/works.js';
import { homePage } from '#lib/content/pages/home.js';

export const load: PageServerLoad = () => {
  const works = getFeaturedWorks(4).map(toWorkListItem);

  return { works, home: homePage };
};