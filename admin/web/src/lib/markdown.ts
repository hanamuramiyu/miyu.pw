import { marked } from 'marked';

const BLOCKED_ELEMENTS = new Set([
  'SCRIPT',
  'IFRAME',
  'OBJECT',
  'EMBED',
  'FORM',
  'INPUT',
  'BUTTON',
  'TEXTAREA',
  'SELECT',
  'STYLE',
  'LINK',
  'META',
  'BASE'
]);

const URL_ATTRIBUTES = new Set(['href', 'src', 'xlink:href']);
const DATA_IMAGE = /^data:image\/(png|jpe?g|gif|webp|avif);/i;
const HEADING_ID = /[ \t]*\{#([a-z0-9]+(?:-[a-z0-9]+)*)\}[ \t]*$/;
const HEADING_TAGS = new Set(['H1', 'H2', 'H3', 'H4', 'H5', 'H6']);

function isSafeUrl(tag: string, name: string, value: string): boolean {
  const lower = value.trim().toLowerCase();

  if (lower.startsWith('javascript:') || lower.startsWith('vbscript:')) {
    return false;
  }
  if (lower.startsWith('data:')) {
    return tag === 'IMG' && (name === 'src' || name === 'xlink:href') && DATA_IMAGE.test(value.trim());
  }
  return true;
}

export function renderMarkdown(source: string): string {
  const html = marked.parse(source, { async: false, gfm: true, breaks: false });
  if (html === '') return '';

  const parsed = new DOMParser().parseFromString(
    `<div id="markdown-root">${html}</div>`,
    'text/html'
  );
  const root = parsed.getElementById('markdown-root');
  if (!root) return '';

  for (const element of Array.from(root.querySelectorAll('*'))) {
    const tag = element.tagName.toUpperCase();

    if (BLOCKED_ELEMENTS.has(tag)) {
      element.remove();
      continue;
    }

    for (const attribute of Array.from(element.attributes)) {
      const name = attribute.name.toLowerCase();

      if (name.startsWith('on') || name === 'style') {
        element.removeAttribute(attribute.name);
        continue;
      }
      if (URL_ATTRIBUTES.has(name) && !isSafeUrl(tag, name, attribute.value)) {
        element.removeAttribute(attribute.name);
      }
    }

    if (tag === 'A') {
      element.setAttribute('target', '_blank');
      element.setAttribute('rel', 'noopener noreferrer');
    }

    if (HEADING_TAGS.has(tag)) {
      applyHeadingId(element);
    }
  }

  return root.innerHTML;
}

function applyHeadingId(heading: Element): void {
  const nodes = Array.from(heading.childNodes);
  for (let index = nodes.length - 1; index >= 0; index -= 1) {
    const node = nodes[index];
    if (node.nodeType !== Node.TEXT_NODE) continue;

    const text = node.textContent ?? '';
    const marker = HEADING_ID.exec(text);
    if (!marker) continue;

    node.textContent = text.replace(HEADING_ID, '');
    if (!heading.id) heading.id = marker[1];
    return;
  }
}
