# gomock

A generic, abstract mock-entity web API written in Go — define your own entity
types and data, and it exposes them as a fully-featured REST resource. Drop a
JSON file in the `data/` directory and it instantly gets listing, pagination,
lookup, and filtering — no code changes required. The bundled sample data
represents Marvel Studios characters, but the API itself has no idea what a
"character" is — swap in any dataset you like.

## Concept

Every record served by the API is normalized into a generic envelope:

```json
{
  "id": 1,
  "name": "Iron Man",
  "slug": "iron-man",
  "attributes": {
    "realName": "Tony Stark",
    "affiliation": "Avengers",
    "actor": "Robert Downey Jr."
  }
}
```

- `id` and `name` come straight from your source JSON (both **required**).
- `slug` is used for lookups; it's taken from your source JSON if present,
  otherwise auto-generated from `name` (e.g. `"Scarlet Witch"` -> `scarlet-witch`).
- Every other field you supply lands in `attributes` untouched, so each entity
  type can have a completely different shape.

## Adding your own entity type

1. Create `data/<yourtype>.json` containing a JSON array of flat objects:

   ```json
   [
     { "id": 1, "name": "Sword", "damage": 12, "rarity": "common" },
     { "id": 2, "name": "Shield", "defense": 8, "rarity": "rare" }
   ]
   ```

2. Start (or restart) the server. The filename (without `.json`) becomes both
   the type name and the route prefix, e.g. `data/weapons.json` is served at
   `/api/v1/weapons`.

Every record **must** include `id` (number) and `name` (string); the server
fails fast at startup with a descriptive error (file + record index) if either
is missing.

## Running locally

```sh
go run ./cmd/server
```

Configuration is via environment variables:

| Variable  | Default   | Description                              |
|-----------|-----------|-------------------------------------------|
| `PORT`    | `8080`    | TCP port the HTTP server listens on       |
| `DATA_DIR`| `./data`  | Directory scanned for `*.json` type files |

## Web console

Open <http://localhost:8080/> in a browser and you get an interactive console
instead of raw JSON. It discovers your entity types from the live API, so it
works with whatever data you dropped in `DATA_DIR` — no configuration.

Pick a type, set `limit`/`offset`, add filters (the key box suggests the exact
attribute spellings found on your records), and hit **Send request**. The
response panel shows the status, timing, body size and pretty-printed JSON;
`Prev`/`Next` follow the pagination links, and clicking a result row fetches
that single entity. There is also **Copy as curl** for taking a request you
built to the terminal.

The console is compiled into the binary, loads nothing from the network, and
adds no dependencies — the whole page is one embedded HTML file.

## API reference

All routes are prefixed with `/api/v1`.

### `GET /`
Content-negotiated. Clients whose `Accept` header explicitly prefers
`text/html` — i.e. browsers — get the [web console](#web-console). Everybody
else, including `curl` and any client sending `Accept: */*`, gets the JSON
discovery document: the API name plus a link for every currently loaded entity
type, so clients can explore the API without prior knowledge of its data.

```sh
curl localhost:8080/                          # JSON discovery document
curl -H 'Accept: text/html' localhost:8080/   # HTML console
```

`?format=json` forces JSON, which is how you read the discovery document from
a browser.

### `GET /console`
The web console, served unconditionally regardless of `Accept`. Useful as a
stable link to share.

### `GET /api/v1/types`
List every discovered entity type and its record count.

```sh
curl localhost:8080/api/v1/types
```

### `GET /api/v1/{type}`
List entities of a type, paginated and optionally filtered.

Query parameters:
- `limit` (default `20`, max `100`)
- `offset` (default `0`)
- any other query parameter is treated as an equality filter against a core
  field (`name`, `slug`) or an `attributes` field, e.g. `?affiliation=Avengers`

Filter semantics worth knowing:
- Attribute keys are **case-sensitive** (`realName`, not `realname`); only
  `name` and `slug` are matched case-insensitively.
- Values are compared case-insensitively but must match exactly — there is no
  partial match, no full-text search, and no sorting.
- An unrecognised key returns **zero results**, not an error.
- Array-valued attributes (e.g. `powers`) are stringified as
  `[flight telekinesis]`, so they are effectively unfilterable.

```sh
curl "localhost:8080/api/v1/characters?limit=2&offset=0"
curl "localhost:8080/api/v1/characters?affiliation=Avengers"
```

Response shape:

```json
{
  "count": 64,
  "next": "http://localhost:8080/api/v1/characters?limit=2&offset=2",
  "previous": null,
  "results": [ { "id": 1, "name": "Iron Man", "slug": "iron-man", "attributes": { } } ]
}
```

`next` and `previous` are absolute URLs. They honour `X-Forwarded-Proto`, so
they come back as `https://` behind a TLS-terminating proxy such as Render's.

### `GET /api/v1/{type}/{idOrSlug}`
Fetch a single entity by numeric ID or by slug.

```sh
curl localhost:8080/api/v1/characters/1
curl localhost:8080/api/v1/characters/iron-man
```

### `PUT /api/v1/admin/{type}`
Replace the entire in-memory dataset for a type with a new JSON array (same
flat-object format as the data files). This also works for a type name that
doesn't exist yet, effectively creating it at runtime. **Changes made this way
are not persisted to disk** — restarting the server reverts to the JSON files
in `DATA_DIR`.

```sh
curl -X PUT localhost:8080/api/v1/admin/items \
  -H 'Content-Type: application/json' \
  -d '[{"id":1,"name":"Elixir","potency":"high"}]'
```

## Docker

```sh
docker build -t gomock .
docker run -p 8080:8080 gomock
```

To use your own custom data instead of the bundled samples, mount a volume
over `/app/data`:

```sh
docker run -p 8080:8080 -v "$(pwd)/mydata:/app/data" gomock
```

## Deploying to Render (free tier)

A [render.yaml](render.yaml) Blueprint is included so the service definition
lives in source control:

1. Push this repo to GitHub (already done if you're reading this from there).
2. In the [Render dashboard](https://dashboard.render.com/), choose
   **New > Blueprint** and connect this repository. Render will detect
   `render.yaml` and provision a free `docker` web service automatically.
3. Render injects its own `PORT` environment variable at runtime (overriding
   the Dockerfile default), so no extra configuration is needed.

Alternatively, without the Blueprint: **New > Web Service** > connect the
repo > Environment: `Docker` > Instance Type: `Free`.

Note: Render's free tier spins the service down after a period of
inactivity, so the first request after idling will be slow (cold start).

## Notes

- No authentication is enforced on the admin override endpoint — this project
  is intended for local/dev mock use. Add middleware in front of it if you
  expose it beyond that. For the same reason the web console documents that
  endpoint but deliberately provides no button that calls it.
- The project has no third-party dependencies; `go.mod` lists none and there
  is no `go.sum`. The console is plain embedded HTML for the same reason.
