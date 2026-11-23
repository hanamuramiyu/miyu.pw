import matter from 'gray-matter';
import { Marked, Renderer, type Tokens } from 'marked';
import {
  WorkFrontmatterSchema,
  type Work,
  type WorkHeading,
  type WorkListItem
} from './schema';

const workModules = import.meta.glob('/content/works/*/index.md', {
  eager: true,
  query: '?raw',
  import: 'default'
}) as Record<string, string>;

const slugPattern = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
const keyPattern = /^\/content\/works\/([^/]+)\/index\.md$/;

function contentError(file: string, message: string): never {
  throw new Error(`[content] ${file}: ${message}`);
}

function headingLabel(text: string): string {
  return text
    .replace(/!\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/[*_~`]/g, '')
    .trim();
}

function slugify(label: string): string {
  const slug = label
    .toLowerCase()
    .normalize('NFKD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');

  return slug || 'section';
}

const headingLinePattern = /^(#{1,6})[ \t]+(.*)$/;
const explicitIdPattern = /[ \t]*\{#([a-z0-9]+(?:-[a-z0-9]+)*)\}[ \t]*$/;
const headingIdMarker = '<!--work-heading-id:';
const headingIdMarkerPattern = /\s*<!--work-heading-id:([a-z0-9]+(?:-[a-z0-9]+)*)-->$/;
const fencePattern = /^[ \t]{0,3}(`{3,}|~{3,})/;

function prepareBody(body: string): string {
  let fence: string | null = null;

  return body
    .split('\n')
    .map((line) => {
      const fenceMatch = line.match(fencePattern);
      if (fenceMatch) {
        const marker = fenceMatch[1][0];
        if (fence === null) {
          fence = marker;
        } else if (fence === marker) {
          fence = null;
        }
        return line;
      }

      if (fence !== null) {
        return line;
      }

      const headingMatch = line.match(headingLinePattern);
      if (!headingMatch) {
        return line;
      }

      const idMatch = headingMatch[2].match(explicitIdPattern);
      if (!idMatch || idMatch.index === undefined) {
        return line;
      }

      const label = headingMatch[2].slice(0, idMatch.index).trimEnd();
      return `${headingMatch[1]} ${label} ${headingIdMarker}${idMatch[1]}-->`;
    })
    .join('\n');
}

function renderBody(body: string): { html: string; headings: WorkHeading[] } {
  const headings: WorkHeading[] = [];
  const usedIds = new Set<string>();
  const renderer = new Renderer();

  renderer.heading = function ({ tokens, depth, text }: Tokens.Heading) {
    const idMatch = text.match(headingIdMarkerPattern);
    const explicitId = idMatch?.[1] ?? null;
    const cleanText = explicitId ? text.replace(headingIdMarkerPattern, '') : text;
    const inlineTokens = explicitId
      ? tokens.filter(
          (token) =>
            !(token.type === 'html' && token.raw.trim().startsWith(headingIdMarker))
        )
      : tokens;
    const inline = this.parser.parseInline(inlineTokens);

    if (depth !== 2) {
      return `<h${depth}>${inline}</h${depth}>\n`;
    }

    const label = headingLabel(cleanText);
    const base = explicitId ?? slugify(label);
    let id = base;
    let suffix = 2;

    if (explicitId && usedIds.has(id)) {
      throw new Error(`[content] Duplicate heading id "${id}".`);
    }

    while (usedIds.has(id)) {
      id = `${base}-${suffix}`;
      suffix += 1;
    }

    usedIds.add(id);
    headings.push({ id, label });

    return `<h2 id="${id}">${inline}</h2>\n`;
  };

  const parser = new Marked({ renderer });
  const html = parser.parse(prepareBody(body), { async: false });

  return { html, headings };
}

function parseWork(file: string, source: string): Work {
  const slug = file.match(keyPattern)?.[1];
  if (!slug) {
    contentError(file, 'Expected the path /content/works/<slug>/index.md');
  }
  if (!slugPattern.test(slug)) {
    contentError(file, `Invalid slug "${slug}". Use lowercase letters, digits and single hyphens.`);
  }

  const parsed = matter(source);
  const result = WorkFrontmatterSchema.safeParse(parsed.data);
  if (!result.success) {
    const issues = result.error.issues
      .map((issue) => `${issue.path.join('.') || '(frontmatter)'}: ${issue.message}`)
      .join('; ');
    contentError(file, issues);
  }

  const body = parsed.content.trim();
  if (!body) {
    contentError(file, 'The Markdown body is empty.');
  }

  const frontmatter = result.data;
  const rendered = renderBody(body);

  return {
    ...frontmatter,
    slug,
    bodyHtml: rendered.html,
    headings: rendered.headings,
    updated: frontmatter.updated ?? frontmatter.date,
    featured: frontmatter.featured ?? false,
    published: frontmatter.published ?? true,
    links: frontmatter.links ?? [],
    images: frontmatter.images ?? []
  };
}

const allWorks = Object.entries(workModules)
  .map(([file, source]) => parseWork(file, source))
  .sort((a, b) => b.date.localeCompare(a.date));

const publishedWorks = allWorks.filter((work) => work.published);

export function getAllWorks(): Work[] {
  return publishedWorks;
}

export function getFeaturedWorks(limit = 4): Work[] {
  return publishedWorks
    .filter((work) => work.featured)
    .sort((a, b) => {
      const orderA = a.order ?? Number.POSITIVE_INFINITY;
      const orderB = b.order ?? Number.POSITIVE_INFINITY;
      return orderA === orderB ? b.date.localeCompare(a.date) : orderA - orderB;
    })
    .slice(0, limit);
}

export function getWork(slug: string): Work {
  const work = publishedWorks.find((item) => item.slug === slug);
  if (!work) {
    throw new Error(`[content] No published work found for slug "${slug}".`);
  }
  return work;
}

export function getWorkSlugs(): string[] {
  return publishedWorks.map((work) => work.slug);
}

export function toWorkListItem(work: Work): WorkListItem {
  return {
    slug: work.slug,
    title: work.title,
    summary: work.summary,
    date: work.date,
    updated: work.updated,
    tools: work.tools,
    links: work.links,
    published: work.published
  };
}