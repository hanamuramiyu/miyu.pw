import type { RequestHandler } from './$types';
import { site } from '#lib/data/site.js';
import { getAllWorks } from '#lib/content/works.js';

export const prerender = true;

interface SitemapRoute {
  path: string;
  lastmod?: string;
}

const routes: SitemapRoute[] = [
  { path: '/' },
  { path: '/works/' },
  { path: '/about/' },
  { path: '/process/' },
  { path: '/design/' },
  ...getAllWorks().map((work) => ({
    path: `/works/${work.slug}/`,
    lastmod: work.updated
  }))
];

function escapeXml(value: string): string {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&apos;');
}

function absoluteUrl(pathname = '/'): string {
  const base = site.url.replace(/\/+$/, '');
  const path = pathname.replace(/^\/+/, '').replace(/\/+$/, '');

  return path ? `${base}/${path}/` : `${base}/`;
}

function renderUrl(route: SitemapRoute): string {
  const loc = `<loc>${escapeXml(absoluteUrl(route.path))}</loc>`;
  const lastmod = route.lastmod ? `<lastmod>${escapeXml(route.lastmod)}</lastmod>` : '';

  return `  <url>${loc}${lastmod}</url>`;
}

export const GET: RequestHandler = () => {
  const xml = [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">',
    ...routes.map(renderUrl),
    '</urlset>',
    ''
  ].join('\n');

  return new Response(xml, {
    headers: { 'Content-Type': 'application/xml; charset=utf-8' }
  });
};