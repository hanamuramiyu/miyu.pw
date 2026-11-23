import type { PageServerLoad } from './$types';
import { processPage } from '#lib/content/pages/process.js';

export const load: PageServerLoad = () => {
  return { process: processPage };
};