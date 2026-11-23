import type { PageServerLoad } from './$types';
import { getAllWorks, toWorkListItem } from '#lib/content/works.js';
import { worksPage } from '#lib/content/pages/works.js';

export const load: PageServerLoad = () => {
  return {
    works: getAllWorks().map(toWorkListItem),
    worksPage
  };
};
