import type { PageServerLoad } from './$types';
import { homePage } from '#lib/content/pages/home.js';

export const load: PageServerLoad = () => {
  return { about: homePage };
};
