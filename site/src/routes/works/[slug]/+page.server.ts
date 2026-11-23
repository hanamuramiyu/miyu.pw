import type { EntryGenerator, PageServerLoad } from './$types';
import { error } from '@sveltejs/kit';
import { getWork, getWorkSlugs } from '#lib/content/works.js';

export const entries: EntryGenerator = () => getWorkSlugs().map((slug) => ({ slug }));

export const load: PageServerLoad = ({ params }) => {
  try {
    return { work: getWork(params.slug) };
  } catch {
    error(404, 'Work not found.');
  }
};
