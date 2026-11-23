import type { RequestHandler } from './$types';
import { site } from '#lib/data/site.js';

export const prerender = true;

export const GET: RequestHandler = () => {
  const lines = site.noindex
    ? ['User-agent: *', 'Disallow: /']
    : ['User-agent: *', 'Allow: /', `Sitemap: ${site.url}/sitemap.xml`];

  return new Response(`${lines.join('\n')}\n`, {
    headers: { 'Content-Type': 'text/plain; charset=utf-8' }
  });
};
