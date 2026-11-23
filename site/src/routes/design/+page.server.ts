import type { PageServerLoad } from './$types';
import { designPage } from '#lib/content/pages/design.js';

export const load: PageServerLoad = () => {
  return { design: designPage };
};