import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api, type JobRun } from '../lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { Badge } from '../components/ui/Badge'
import { formatDate } from '../lib/utils'

const STAGE_ORDER = [
  'research_collect', 'research_filter', 'brief_generate',
  'script_outline_generate', 'script_generate', 'script_rewrite_for_audio',
  'voice_render_segments', 'audio_assemble', 'transcript_finalize',
  'metadata_generate', 'video_package', 'publish_prepare', 'publish_deliver',
]

const STATUS_VARIANT: Record<string, 'default' | 'success' | 'warning' | 'error' | 'info'> = {
  draft: 'default',
  queued: 'info',
  researching: 'info',
  brief_ready: 'info',
  scripting: 'info',
  voice_rendering: 'info',
  rendering: 'info',
  ready_for_review: 'warning',
  scheduled: 'warning',
  published: 'success',
  failed: 'error',
}

const JOB_STATUS_VARIANT: Record<string, 'default' | 'success' | 'warning' | 'error' | 'info'> = {
  pending: 'default',
  running: 'info',
  completed: 'success',
  failed: 'error',
  retrying: 'warning',
}

function PipelineProgress({ jobs }: { jobs: JobRun[] }) {
  const jobMap: Record<string, JobRun> = {}
  for (const j of jobs) {
    jobMap[j.stage] = j
  }
  return (
    <div className="overflow-x-auto">
      <div className="flex gap-1 min-w-max py-2">
        {STAGE_ORDER.map((stage, i) => {
          const job = jobMap[stage]
          const status = job?.status ?? 'pending'
          const bg =
            status === 'completed' ? 'bg-green-500' :
              status === 'running' ? 'bg-blue-500 animate-pulse' :
                status === 'failed' ? 'bg-red-500' :
                  'bg-gray-200'
          return (
            <div key={stage} className="flex flex-col items-center gap-1">
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-xs text-white font-medium ${bg}`}>
                {i + 1}
              </div>
              <span className="text-xs text-gray-500 text-center w-16 leading-tight">
                {stage.replace(/_/g, ' ')}
              </span>
            </div>
          )
        })}
      </div>
    </div>
  )
}

export default function EpisodeDetail() {
  const { episodeID } = useParams<{ episodeID: string }>()
  const qc = useQueryClient()
  const [urlForm, setUrlForm] = useState({ url: '', title: '' })
  const [textForm, setTextForm] = useState({ title: '', author: '', text: '' })

  const { data: episode, isLoading } = useQuery({
    queryKey: ['episode', episodeID],
    queryFn: () => api.getEpisode(episodeID!),
    enabled: !!episodeID,
    refetchInterval: 5000,
  })

  const { data: jobs = [] } = useQuery({
    queryKey: ['jobs', episodeID],
    queryFn: () => api.listJobs(episodeID!),
    enabled: !!episodeID,
    refetchInterval: 5000,
  })

  const { data: sources = [] } = useQuery({
    queryKey: ['sources', episodeID],
    queryFn: () => api.listSources(episodeID!),
    enabled: !!episodeID,
  })

  const { data: script } = useQuery({
    queryKey: ['script', episodeID],
    queryFn: () => api.getLatestScript(episodeID!),
    enabled: !!episodeID,
    retry: false,
  })

  const queueMutation = useMutation({
    mutationFn: () => api.queueEpisode(episodeID!),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['episode', episodeID] }),
  })

  const retryMutation = useMutation({
    mutationFn: () => api.retryEpisode(episodeID!),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['episode', episodeID] }),
  })

  const importURLMutation = useMutation({
    mutationFn: () => api.importSourceURL(episodeID!, urlForm.url, urlForm.title),
    onSuccess: () => {
      setUrlForm({ url: '', title: '' })
      qc.invalidateQueries({ queryKey: ['sources', episodeID] })
    },
  })

  const importTextMutation = useMutation({
    mutationFn: () => api.importSourceText(episodeID!, textForm.text, textForm.title, textForm.author),
    onSuccess: () => {
      setTextForm({ title: '', author: '', text: '' })
      qc.invalidateQueries({ queryKey: ['sources', episodeID] })
    },
  })

  if (isLoading) return <p className="text-gray-400">Loading…</p>
  if (!episode) return <p className="text-red-500">Episode not found.</p>

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Link to="/episodes" className="text-gray-400 hover:text-gray-600 text-sm">← Episodes</Link>
        <span className="text-gray-300">/</span>
        <h1 className="text-xl font-bold text-gray-900">{episode.title || episode.topic || 'Untitled'}</h1>
        <Badge variant={STATUS_VARIANT[episode.status] ?? 'default'}>
          {episode.status.replace(/_/g, ' ')}
        </Badge>
      </div>

      <div className="flex gap-3">
        {['draft', 'brief_ready', 'ready_for_review'].includes(episode.status) && (
          <Button onClick={() => queueMutation.mutate()} disabled={queueMutation.isPending}>
            {queueMutation.isPending ? 'Queueing…' : '▶ Queue Pipeline'}
          </Button>
        )}
        {episode.status === 'failed' && (
          <Button variant="danger" onClick={() => retryMutation.mutate()} disabled={retryMutation.isPending}>
            {retryMutation.isPending ? 'Retrying…' : '↻ Retry'}
          </Button>
        )}
      </div>

      {episode.error_message && (
        <div className="bg-red-50 border border-red-200 rounded-lg px-4 py-3 text-sm text-red-700">
          <strong>Error:</strong> {episode.error_message}
        </div>
      )}

      <div className="grid md:grid-cols-2 gap-6">
        <Card>
          <CardHeader><CardTitle>Details</CardTitle></CardHeader>
          <CardContent>
            <dl className="space-y-3 text-sm">
              <div className="flex justify-between">
                <dt className="text-gray-500">Topic</dt>
                <dd className="text-gray-900 font-medium">{episode.topic || '—'}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">Angle</dt>
                <dd className="text-gray-900 max-w-xs text-right">{episode.angle || '—'}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">Target Duration</dt>
                <dd className="text-gray-900">{Math.round(episode.target_duration_seconds / 60)} min</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">Created</dt>
                <dd className="text-gray-900">{formatDate(episode.created_at)}</dd>
              </div>
              {episode.published_at && (
                <div className="flex justify-between">
                  <dt className="text-gray-500">Published</dt>
                  <dd className="text-gray-900">{formatDate(episode.published_at)}</dd>
                </div>
              )}
            </dl>
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle>Pipeline Progress</CardTitle></CardHeader>
          <CardContent>
            <PipelineProgress jobs={jobs} />
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Add / Import Sources</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid md:grid-cols-2 gap-6">
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Source URL</label>
                <input
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={urlForm.url}
                  onChange={(e) => setUrlForm((f) => ({ ...f, url: e.target.value }))}
                  placeholder="https://example.com/article"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Title (optional)</label>
                <input
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={urlForm.title}
                  onChange={(e) => setUrlForm((f) => ({ ...f, title: e.target.value }))}
                  placeholder="Article or video title"
                />
              </div>
              <Button
                onClick={() => importURLMutation.mutate()}
                disabled={importURLMutation.isPending || !urlForm.url.trim()}
              >
                {importURLMutation.isPending ? 'Importing URL…' : 'Import URL'}
              </Button>
              {importURLMutation.error && (
                <p className="text-sm text-red-600">{(importURLMutation.error as Error).message}</p>
              )}
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Paste text / notes</label>
                <textarea
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  rows={6}
                  value={textForm.text}
                  onChange={(e) => setTextForm((f) => ({ ...f, text: e.target.value }))}
                  placeholder="Paste transcript, research notes, or reference text here"
                />
              </div>
              <div className="grid sm:grid-cols-2 gap-3">
                <input
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={textForm.title}
                  onChange={(e) => setTextForm((f) => ({ ...f, title: e.target.value }))}
                  placeholder="Title (optional)"
                />
                <input
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={textForm.author}
                  onChange={(e) => setTextForm((f) => ({ ...f, author: e.target.value }))}
                  placeholder="Author (optional)"
                />
              </div>
              <Button
                variant="secondary"
                onClick={() => importTextMutation.mutate()}
                disabled={importTextMutation.isPending || !textForm.text.trim()}
              >
                {importTextMutation.isPending ? 'Saving text…' : 'Add Pasted Text'}
              </Button>
              {importTextMutation.error && (
                <p className="text-sm text-red-600">{(importTextMutation.error as Error).message}</p>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid md:grid-cols-2 gap-6">
        <Card>
          <CardHeader><CardTitle>Sources ({sources.length})</CardTitle></CardHeader>
          <CardContent className="p-0">
            {sources.length === 0 ? (
              <p className="p-6 text-gray-400 text-sm">No sources added yet. Use the form above to import a URL or paste text.</p>
            ) : (
              <ul className="divide-y divide-gray-100">
                {sources.map((src) => (
                  <li key={src.id} className="px-6 py-3">
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0">
                        <p className="text-sm font-medium text-gray-900 truncate">
                          {src.source_title || src.source_url || 'Untitled source'}
                        </p>
                        {src.source_url && (
                          <a href={src.source_url} target="_blank" rel="noreferrer"
                            className="text-xs text-indigo-600 hover:underline truncate block">
                            {src.source_url}
                          </a>
                        )}
                        {!src.source_url && src.extracted_text && (
                          <p className="text-xs text-gray-500 mt-1 line-clamp-2">
                            {src.extracted_text}
                          </p>
                        )}
                      </div>
                      <Badge variant={src.selected ? 'success' : 'default'}>
                        {src.selected ? 'selected' : 'unselected'}
                      </Badge>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle>Job Runs ({jobs.length})</CardTitle></CardHeader>
          <CardContent className="p-0">
            {jobs.length === 0 ? (
              <p className="p-6 text-gray-400 text-sm">No jobs yet.</p>
            ) : (
              <ul className="divide-y divide-gray-100">
                {jobs.map((j: JobRun) => (
                  <li key={j.id} className="px-6 py-3">
                    <div className="flex items-center justify-between">
                      <div>
                        <span className="text-sm font-medium text-gray-900">
                          {j.stage.replace(/_/g, ' ')}
                        </span>
                        <span className="text-xs text-gray-400 ml-2">attempt {j.attempt}</span>
                      </div>
                      <Badge variant={JOB_STATUS_VARIANT[j.status] ?? 'default'}>{j.status}</Badge>
                    </div>
                    {j.error_message && (
                      <p className="text-xs text-red-500 mt-1">{j.error_message}</p>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      </div>

      {script && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>Script v{script.version}</CardTitle>
              <Badge variant={script.status === 'approved' ? 'success' : 'default'}>{script.status}</Badge>
            </div>
          </CardHeader>
          <CardContent>
            <pre className="text-sm text-gray-700 whitespace-pre-wrap font-sans leading-relaxed max-h-96 overflow-y-auto">
              {script.spoken_text || script.full_text || 'No script content.'}
            </pre>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
