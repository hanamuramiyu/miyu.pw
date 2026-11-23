import { z } from 'zod';

const text = (max: number) => z.string().trim().min(1).max(max);

const isoDate = z
  .string()
  .trim()
  .regex(/^\d{4}-\d{2}-\d{2}$/, 'Expected an ISO date in YYYY-MM-DD form')
  .refine((value) => {
    const date = new Date(`${value}T00:00:00Z`);
    return !Number.isNaN(date.getTime()) && date.toISOString().slice(0, 10) === value;
  }, 'Expected a real calendar date');

const fileName = z
  .string()
  .trim()
  .min(1)
  .max(160)
  .refine((value) => !value.includes('/') && !value.includes('..'), {
    message: 'Expected a file name inside the work directory'
  });

const imageRef = z.strictObject({
  src: fileName,
  alt: text(300)
});

const mediaRef = z
  .string()
  .trim()
  .min(1)
  .max(300)
  .refine((value) => value.startsWith('/media/') && !value.includes('..'), {
    message: 'Expected a path inside /media/'
  });

const workImage = z.strictObject({
  src: mediaRef,
  alt: text(300),
  caption: text(200).optional(),
  width: z.number().int().positive(),
  height: z.number().int().positive(),
  cover: z.boolean().optional()
});

const linkRef = z.strictObject({
  label: text(60),
  href: z.string().trim().regex(/^https?:\/\//, 'Expected an http(s) URL')
});

const colorRef = z.strictObject({
  name: text(60),
  hex: z.string().trim().regex(/^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$/, 'Expected a hex colour')
});

const fontRef = z.strictObject({
  name: text(80),
  role: text(80)
});

const seoRef = z.strictObject({
  title: text(120).optional(),
  description: text(200).optional(),
  image: z.string().trim().regex(/^\//, 'Expected a public path beginning with /').optional()
});

export const WorkFrontmatterSchema = z.strictObject({
  title: text(120),
  summary: text(200),
  date: isoDate,
  updated: isoDate.optional(),
  tools: z.array(text(60)).min(1).max(12),
  featured: z.boolean().optional(),
  order: z.number().int().min(0).optional(),
  published: z.boolean().optional(),
  client: text(120).optional(),
  role: text(120).optional(),
  kind: text(80).optional(),
  audience: text(120).optional(),
  timeline: text(120).optional(),
  cover: imageRef.optional(),
  links: z.array(linkRef).max(8).optional(),
  colors: z.array(colorRef).max(12).optional(),
  fonts: z.array(fontRef).max(8).optional(),
  images: z.array(workImage).optional(),
  gallery: z.array(imageRef.extend({ caption: text(200).optional() })).max(24).optional(),
  seo: seoRef.optional()
});

export type WorkFrontmatter = z.infer<typeof WorkFrontmatterSchema>;
export type WorkLink = z.infer<typeof linkRef>;
export type WorkImage = z.infer<typeof workImage>;

export interface WorkHeading {
  id: string;
  label: string;
}

export interface Work extends WorkFrontmatter {
  slug: string;
  bodyHtml: string;
  headings: WorkHeading[];
  updated: string;
  featured: boolean;
  published: boolean;
  links: WorkLink[];
  images: WorkImage[];
}

export interface WorkListItem {
  slug: string;
  title: string;
  summary: string;
  date: string;
  updated: string;
  tools: string[];
  links: WorkLink[];
  published: boolean;
}