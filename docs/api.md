# PodGen Network API

Base URL: `http://localhost:8080/api/v1`

## Health

```
GET /health
```

## Shows

| Method | Path | Description |
|--------|------|-------------|
| GET | /shows | List all shows |
| POST | /shows | Create a show |
| GET | /shows/{showID} | Get a show |
| PUT | /shows/{showID} | Update a show |
| GET | /shows/{showID}/hosts | List host profiles |
| POST | /shows/{showID}/hosts | Create host profile |
| GET | /shows/{showID}/hosts/{hostID} | Get host profile |

### Create Show

```json
POST /shows
{
  "name": "Tech Weekly",
  "description": "Weekly technology news",
  "language": "en",
  "niche": "technology",
  "tone": "informative",
  "cadence": "weekly",
  "default_duration_minutes": 30,
  "default_format": "solo"
}
```

## Episodes

| Method | Path | Description |
|--------|------|-------------|
| GET | /episodes | List all episodes (query: show_id, status) |
| GET | /shows/{showID}/episodes | List episodes for show |
| POST | /shows/{showID}/episodes | Create episode |
| GET | /episodes/{episodeID} | Get episode |
| PUT | /episodes/{episodeID} | Update episode |
| POST | /episodes/{episodeID}/queue | Queue pipeline |
| POST | /episodes/{episodeID}/retry | Retry failed episode |

### Create Episode

```json
POST /shows/{showID}/episodes
{
  "topic": "The Future of AI",
  "angle": "How AI is transforming creative industries",
  "target_duration_seconds": 1800
}
```

## Sources

| Method | Path | Description |
|--------|------|-------------|
| GET | /episodes/{episodeID}/sources | List sources |
| POST | /episodes/{episodeID}/sources/import-url | Import URL |
| POST | /episodes/{episodeID}/sources/import-text | Import text |
| GET | /episodes/{episodeID}/sources/{sourceID} | Get source |

## Briefs

| Method | Path | Description |
|--------|------|-------------|
| GET | /episodes/{episodeID}/brief | Get brief |
| POST | /episodes/{episodeID}/brief/generate | Generate brief |

## Scripts

| Method | Path | Description |
|--------|------|-------------|
| GET | /episodes/{episodeID}/scripts | List drafts |
| GET | /episodes/{episodeID}/scripts/latest | Get latest draft |
| POST | /episodes/{episodeID}/scripts/generate | Generate script |
| POST | /episodes/{episodeID}/scripts/{draftID}/approve | Approve draft |

## Assets

| Method | Path | Description |
|--------|------|-------------|
| GET | /episodes/{episodeID}/assets | List assets |
| GET | /episodes/{episodeID}/assets/{assetID} | Get asset |
| POST | /episodes/{episodeID}/assets/render-audio | Queue audio render |
| POST | /episodes/{episodeID}/assets/render-video | Queue video render |

## Jobs

| Method | Path | Description |
|--------|------|-------------|
| GET | /episodes/{episodeID}/jobs | List job runs |
| GET | /episodes/{episodeID}/jobs/{jobID} | Get job run |

## Publishing

| Method | Path | Description |
|--------|------|-------------|
| POST | /episodes/{episodeID}/publish | Publish episode |
| GET | /episodes/{episodeID}/publish/status | Get publish status |

## Episode Status Flow

```
draft → queued → researching → brief_ready → scripting
  → voice_rendering → rendering → ready_for_review
  → scheduled → published
  (any stage can → failed)
```
