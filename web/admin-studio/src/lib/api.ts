const BASE = '/api/v1'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...options?.headers },
    ...options,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || res.statusText)
  }
  return res.json()
}

// Types
export interface Show {
  id: string
  slug: string
  name: string
  description: string
  language: string
  niche: string
  tone: string
  status: string
  cadence: string
  default_duration_minutes: number
  default_format: string
  created_at: string
  updated_at: string
}

export interface Episode {
  id: string
  show_id: string
  status: string
  topic: string
  angle: string
  title: string
  subtitle: string
  description: string
  transcript: string
  target_duration_seconds: number
  planned_publish_at?: string
  published_at?: string
  error_message: string
  created_at: string
  updated_at: string
}

export interface JobRun {
  id: string
  episode_id: string
  stage: string
  status: string
  attempt: number
  started_at?: string
  finished_at?: string
  error_message: string
}

export interface EpisodeSource {
  id: string
  episode_id: string
  source_type: string
  source_url: string
  source_title: string
  source_author: string
  extracted_text: string
  summary: string
  relevance_score: number
  trust_score: number
  selected: boolean
  created_at: string
}

export interface ScriptDraft {
  id: string
  episode_id: string
  version: number
  format: string
  full_text: string
  spoken_text: string
  status: string
  created_at: string
}

// API functions
export const api = {
  // Shows
  listShows: () => request<Show[]>('/shows'),
  getShow: (id: string) => request<Show>(`/shows/${id}`),
  createShow: (data: Partial<Show>) =>
    request<Show>('/shows', { method: 'POST', body: JSON.stringify(data) }),
  updateShow: (id: string, data: Partial<Show>) =>
    request<Show>(`/shows/${id}`, { method: 'PUT', body: JSON.stringify(data) }),

  // Episodes
  listEpisodes: (params?: { show_id?: string; status?: string }) => {
    const qs = new URLSearchParams(params as Record<string, string>).toString()
    return request<Episode[]>(`/episodes${qs ? `?${qs}` : ''}`)
  },
  getEpisode: (id: string) => request<Episode>(`/episodes/${id}`),
  createEpisode: (showId: string, data: Partial<Episode>) =>
    request<Episode>(`/shows/${showId}/episodes`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  updateEpisode: (id: string, data: Partial<Episode>) =>
    request<Episode>(`/episodes/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  queueEpisode: (id: string) =>
    request<{ status: string }>(`/episodes/${id}/queue`, { method: 'POST' }),
  retryEpisode: (id: string) =>
    request<{ status: string }>(`/episodes/${id}/retry`, { method: 'POST' }),

  // Jobs
  listJobs: (episodeId: string) => request<JobRun[]>(`/episodes/${episodeId}/jobs`),

  // Sources
  listSources: (episodeId: string) =>
    request<EpisodeSource[]>(`/episodes/${episodeId}/sources`),
  importSourceURL: (episodeId: string, url: string, title: string) =>
    request<EpisodeSource>(`/episodes/${episodeId}/sources/import-url`, {
      method: 'POST',
      body: JSON.stringify({ url, title }),
    }),

  // Scripts
  listScripts: (episodeId: string) =>
    request<ScriptDraft[]>(`/episodes/${episodeId}/scripts`),
  getLatestScript: (episodeId: string) =>
    request<ScriptDraft>(`/episodes/${episodeId}/scripts/latest`),
}
