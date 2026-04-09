# PodGen Pipeline Architecture

## Overview

The PodGen pipeline is a multi-stage automated system that takes an episode topic
and produces a fully produced podcast episode. Each stage is processed by the
worker service via a Redis queue.

## Pipeline Stages

```
1.  research_collect         - Gather sources from URLs, feeds, and text
2.  research_filter          - Filter and rank sources by relevance
3.  brief_generate           - Use LLM to synthesize a production brief
4.  script_outline_generate  - Create structured episode outline
5.  script_generate          - Generate full script from outline + brief
6.  script_rewrite_for_audio - Rewrite script optimized for TTS
7.  voice_render_segments    - Send spoken text to TTS provider
8.  audio_assemble           - Assemble final MP3 with FFmpeg
9.  transcript_finalize      - Generate final transcript
10. metadata_generate        - Generate SEO-optimized title/description
11. video_package            - Create audiogram video (optional)
12. publish_prepare          - Prepare platform-specific metadata
13. publish_deliver          - Upload and publish to podcast host
```

## Components

### Orchestrator (`internal/orchestrator`)
- Manages episode state machine
- Validates stage prerequisites
- Records `job_runs` with status
- Advances episode status
- Enqueues next stage on completion
- Marks episode `failed` on error

### Queue (`internal/queue/redis`)
- Uses Redis LPUSH/BRPOP for reliable queuing
- Processing list tracks inflight messages
- Messages acknowledged after processing

### Worker Dispatcher (`apps/worker/internal/jobs`)
- Dequeues `job.Payload` messages
- Routes to appropriate pipeline stage handler
- Reports success/failure back to orchestrator

## Episode Status Map

| Status | Triggered By |
|--------|-------------|
| `draft` | Episode created |
| `queued` | POST /queue called |
| `researching` | research_collect or research_filter running |
| `brief_ready` | brief_generate completed (set externally) |
| `scripting` | script stages running |
| `voice_rendering` | voice_render_segments running |
| `rendering` | audio_assemble/metadata running |
| `ready_for_review` | All rendering complete |
| `scheduled` | publish_prepare completed |
| `published` | publish_deliver completed |
| `failed` | Any stage failed |

## Provider Configuration

Configure providers via environment variables:

```bash
LLM_PROVIDER=ollama                    # default local LLM provider
OLLAMA_HOST=http://ollama:11434       # Ollama endpoint
OLLAMA_MODEL=gemma3:latest            # model used for briefs/scripts/metadata

# Optional alternative:
OPENAI_API_KEY=sk-...                  # only used when LLM_PROVIDER=openai
OPENAI_MODEL=gpt-4o-mini

ELEVENLABS_API_KEY=...                 # TTS provider
STORAGE_URL=http://minio:9000          # S3-compatible storage
```

Providers are optional in the worker — if not configured, stages are skipped (no-op). The default worker configuration now uses Ollama for text generation.
