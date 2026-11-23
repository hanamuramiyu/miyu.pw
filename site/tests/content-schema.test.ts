import { describe, expect, test } from 'bun:test';
import { WorkFrontmatterSchema } from '../src/lib/content/schema.ts';

const minimal = {
  title: 'Monban',
  summary: 'Access-control plugin for Minecraft servers.',
  date: '2025-11-07',
  tools: ['Java']
};

describe('WorkFrontmatterSchema', () => {
  test('accepts the required minimum', () => {
    const result = WorkFrontmatterSchema.safeParse(minimal);
    expect(result.success).toBe(true);
  });

  test('rejects unknown top-level fields', () => {
    const result = WorkFrontmatterSchema.safeParse({ ...minimal, tittle: 'typo' });
    expect(result.success).toBe(false);
  });

  test('rejects unknown nested fields', () => {
    const result = WorkFrontmatterSchema.safeParse({
      ...minimal,
      seo: { title: 'Monban', bogus: true }
    });
    expect(result.success).toBe(false);
  });

  test('requires a real ISO calendar date and rejects unquoted YAML dates', () => {
    expect(WorkFrontmatterSchema.safeParse({ ...minimal, date: '2025-02-30' }).success).toBe(false);
    expect(WorkFrontmatterSchema.safeParse({ ...minimal, date: '2025-1-7' }).success).toBe(false);
    expect(WorkFrontmatterSchema.safeParse({ ...minimal, date: new Date() }).success).toBe(false);
    expect(WorkFrontmatterSchema.safeParse({ ...minimal, date: '2024-02-29' }).success).toBe(true);
  });

  test('requires title, summary, date and at least one tool', () => {
    expect(WorkFrontmatterSchema.safeParse({ ...minimal, title: undefined }).success).toBe(false);
    expect(WorkFrontmatterSchema.safeParse({ ...minimal, summary: undefined }).success).toBe(false);
    expect(WorkFrontmatterSchema.safeParse({ ...minimal, date: undefined }).success).toBe(false);
    expect(WorkFrontmatterSchema.safeParse({ ...minimal, tools: [] }).success).toBe(false);
    expect(WorkFrontmatterSchema.safeParse({ ...minimal, tools: [''] }).success).toBe(false);
    expect(
      WorkFrontmatterSchema.safeParse({
        ...minimal,
        tools: Array.from({ length: 13 }, (_, index) => `tool-${index}`)
      }).success
    ).toBe(false);
  });

  test('accepts optional fields and applies no defaults at the schema layer', () => {
    const result = WorkFrontmatterSchema.safeParse({
      ...minimal,
      updated: '2026-10-22',
      featured: true,
      order: 1,
      published: true,
      client: null,
      role: 'Solo developer',
      timeline: '3 weeks'
    });
    expect(result.success).toBe(false);
    expect(
      WorkFrontmatterSchema.safeParse({
        ...minimal,
        updated: '2026-10-22',
        featured: true,
        order: 1,
        published: true,
        role: 'Solo developer',
        timeline: '3 weeks'
      }).success
    ).toBe(true);
  });

  test('validates links and colours', () => {
    expect(
      WorkFrontmatterSchema.safeParse({
        ...minimal,
        links: [{ label: 'Repository', href: 'https://github.com/hanamuramiyu/monban' }]
      }).success
    ).toBe(true);
    expect(
      WorkFrontmatterSchema.safeParse({ ...minimal, links: [{ label: 'x', href: 'ftp://x' }] }).success
    ).toBe(false);
    expect(
      WorkFrontmatterSchema.safeParse({
        ...minimal,
        colors: [{ name: 'Signal', hex: '#ff79b0' }]
      }).success
    ).toBe(true);
    expect(
      WorkFrontmatterSchema.safeParse({ ...minimal, colors: [{ name: 'x', hex: '#12' }] }).success
    ).toBe(false);
  });

  test('requires alt text and safe file names for cover and gallery', () => {
    expect(
      WorkFrontmatterSchema.safeParse({ ...minimal, cover: { src: 'cover.png', alt: 'Cover' } })
        .success
    ).toBe(true);
    expect(
      WorkFrontmatterSchema.safeParse({ ...minimal, cover: { src: 'cover.png' } }).success
    ).toBe(false);
    expect(
      WorkFrontmatterSchema.safeParse({ ...minimal, cover: { src: 'dir/cover.png', alt: 'Cover' } })
        .success
    ).toBe(false);
    expect(
      WorkFrontmatterSchema.safeParse({ ...minimal, cover: { src: '../cover.png', alt: 'Cover' } })
        .success
    ).toBe(false);
    expect(
      WorkFrontmatterSchema.safeParse({
        ...minimal,
        gallery: [{ src: 'shot-01.png', alt: 'Permission editor', caption: 'Role tree' }]
      }).success
    ).toBe(true);
    expect(
      WorkFrontmatterSchema.safeParse({ ...minimal, gallery: [{ src: 'shot-01.png' }] }).success
    ).toBe(false);
  });

  test('enforces summary length and non-negative integer order', () => {
    expect(
      WorkFrontmatterSchema.safeParse({ ...minimal, summary: 'x'.repeat(201) }).success
    ).toBe(false);
    expect(WorkFrontmatterSchema.safeParse({ ...minimal, order: -1 }).success).toBe(false);
    expect(WorkFrontmatterSchema.safeParse({ ...minimal, order: 1.5 }).success).toBe(false);
    expect(WorkFrontmatterSchema.safeParse({ ...minimal, order: 0 }).success).toBe(true);
  });

  test('trims strings and rejects whitespace-only values', () => {
    expect(WorkFrontmatterSchema.safeParse({ ...minimal, title: '   ' }).success).toBe(false);
    expect(WorkFrontmatterSchema.safeParse({ ...minimal, summary: '   ' }).success).toBe(false);
    expect(WorkFrontmatterSchema.safeParse({ ...minimal, tools: ['   '] }).success).toBe(false);
    expect(
      WorkFrontmatterSchema.safeParse({ ...minimal, cover: { src: 'a.png', alt: '   ' } })
        .success
    ).toBe(false);
    expect(
      WorkFrontmatterSchema.safeParse({
        ...minimal,
        gallery: [{ src: 'a.png', alt: 'ok', caption: '   ' }]
      }).success
    ).toBe(false);

    const result = WorkFrontmatterSchema.safeParse({
      ...minimal,
      title: '  Monban  ',
      summary: '  Summary  ',
      date: ' 2025-11-07 ',
      tools: ['  Java  '],
      links: [{ label: ' Repository ', href: ' https://github.com/hanamuramiyu/monban ' }]
    });
    expect(result.success).toBe(true);
    if (!result.success) return;
    expect(result.data.title).toBe('Monban');
    expect(result.data.summary).toBe('Summary');
    expect(result.data.date).toBe('2025-11-07');
    expect(result.data.tools).toEqual(['Java']);
    expect(result.data.links?.[0]?.label).toBe('Repository');
    expect(result.data.links?.[0]?.href).toBe('https://github.com/hanamuramiyu/monban');
  });
});