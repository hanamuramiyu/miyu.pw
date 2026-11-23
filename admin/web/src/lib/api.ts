import type {
  Me,
  PublishStatus,
  Work,
  WorkInput,
  WorkListResponse
} from './types';

export class ApiError extends Error {
  readonly status: number;
  readonly field?: string;

  constructor(message: string, status: number, field?: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.field = field;
  }
}

export interface Notice {
  title: string;
  detail: string;
}

let csrfToken = '';

export function setCSRFToken(token: string): void {
  csrfToken = token;
}

export function errorNotice(error: unknown): Notice {
  if (error instanceof ApiError) {
    if (error.status === 403) return { title: 'Access denied', detail: '' };
    if (error.status === 404) return { title: 'Not found', detail: '' };
    if (error.status === 405) return { title: 'Method not allowed', detail: '' };
    if (error.status >= 500) {
      return { title: 'Server error', detail: 'Something broke on the server. Try again.' };
    }
    return { title: error.message, detail: '' };
  }
  return { title: 'Cannot reach the server.', detail: '' };
}

interface ErrorBody {
  error?: string;
  field?: string;
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const method = (init.method ?? 'GET').toUpperCase();
  const headers = new Headers(init.headers);

  if (
    init.body !== undefined &&
    !(init.body instanceof FormData) &&
    !headers.has('Content-Type')
  ) {
    headers.set('Content-Type', 'application/json');
  }
  if (method !== 'GET' && method !== 'HEAD' && csrfToken !== '') {
    headers.set('X-CSRF-Token', csrfToken);
  }

  const response = await fetch(path, {
    credentials: 'same-origin',
    ...init,
    method,
    headers
  });

  if (response.status === 204) {
    return undefined as T;
  }

  const text = await response.text();
  let body: unknown = null;
  if (text !== '') {
    try {
      body = JSON.parse(text);
    } catch {
      body = null;
    }
  }

  if (!response.ok) {
    const parsed = (body ?? {}) as ErrorBody;
    const message = typeof parsed.error === 'string' && parsed.error !== ''
      ? parsed.error
      : response.statusText || 'Request failed';
    throw new ApiError(message, response.status, parsed.field);
  }

  return body as T;
}

export const api = {
  me(): Promise<Me> {
    return request<Me>('/api/me');
  },

  listWorks(includeDrafts: boolean): Promise<WorkListResponse> {
    const query = includeDrafts ? '?drafts=1' : '';
    return request<WorkListResponse>(`/api/works${query}`);
  },

  getWork(id: string): Promise<Work> {
    return request<Work>(`/api/works/${encodeURIComponent(id)}`);
  },

  createWork(input: WorkInput): Promise<Work> {
    return request<Work>('/api/works', {
      method: 'POST',
      body: JSON.stringify(input)
    });
  },

  updateWork(id: string, input: WorkInput): Promise<Work> {
    return request<Work>(`/api/works/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(input)
    });
  },

  deleteWork(id: string): Promise<void> {
    return request<void>(`/api/works/${encodeURIComponent(id)}`, {
      method: 'DELETE'
    });
  },

  publish(id: string, published: boolean): Promise<void> {
    return request<void>(`/api/works/${encodeURIComponent(id)}/publish`, {
      method: 'POST',
      body: JSON.stringify({ published })
    });
  },

  publishSite(): Promise<{ state: string }> {
    return request<{ state: string }>('/api/publish', { method: 'POST' });
  },

  publishStatus(): Promise<PublishStatus> {
    return request<PublishStatus>('/api/publish/status');
  },

  logout(): Promise<void> {
    return request<void>('/auth/logout', { method: 'POST' });
  }
};
