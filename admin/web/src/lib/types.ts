export interface Link {
  label: string;
  href: string;
}

export interface Color {
  name: string;
  hex: string;
}

export interface Font {
  name: string;
  role: string;
}

export interface Image {
  path: string;
  alt: string;
  caption: string;
  width: number;
  height: number;
  isCover: boolean;
}

export interface Work {
  id: string;
  slug: string;
  title: string;
  summary: string;
  bodyMd: string;
  kind: string;
  audience: string;
  role: string;
  client: string | null;
  date: string;
  updated: string;
  published: boolean;
  featured: boolean;
  sortOrder: number;
  tools: string[];
  links: Link[];
  colors: Color[];
  fonts: Font[];
  images: Image[];
  createdAt: string;
  updatedAt: string;
}

export interface WorkSummary {
  id: string;
  slug: string;
  title: string;
  summary: string;
  kind: string;
  date: string;
  updated: string;
  published: boolean;
  featured: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export interface WorkInput {
  slug: string;
  title: string;
  summary: string;
  bodyMd: string;
  kind: string;
  audience: string;
  role: string;
  client: string | null;
  date: string;
  updated: string;
  published: boolean;
  featured: boolean;
  sortOrder: number;
  tools: string[];
  links: Link[];
  colors: Color[];
  fonts: Font[];
  images: Image[];
}

export interface Me {
  login: string;
  csrfToken: string;
}

export interface PublishStatus {
  state: 'idle' | 'running' | 'ok' | 'failed' | string;
  startedAt?: string;
  finishedAt?: string;
  release?: string;
  error?: string;
}

export interface WorkListResponse {
  works: WorkSummary[];
}
