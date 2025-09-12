# Web Interface Requirements

## GET /projects/:prid/struct

- Supports optional query parameters to limit and filter output:
  - `depth` – maximum depth of nested requirements to include.
  - `status` – filter requirements by status (e.g. active, deleted).
  - `page` – page number for paginated results.
- When provided, the backend filters and paginates the project structure before returning JSON.
- Without any query parameters, the entire project structure is returned.
