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

## API reference

All routes are prefixed with `/api/v1`.

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

## Notes

- No authentication is enforced on the admin override endpoint — this project
  is intended for local/dev mock use. Add middleware in front of it if you
  expose it beyond that.
- No automated test suite is included in this initial version.
