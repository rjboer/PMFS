# Web Interface Requirements

This document captures the requirements for the web interface example.

## Retrieve Project Structure
- The interface must expose an endpoint `GET /projects/:prid/struct`.
- The endpoint returns a nested structure describing the entire project, including:
  - Product data.
  - Project metadata.
  - Requirements with their associated attachments.
- The client fetches this structure and passes it to the DOM for rendering.

## Data Format
- Responses are JSON objects where nested arrays or maps represent hierarchy.
- All fields required by the DOM must be included and use stable names.

## Performance
- Requests for the full project structure should complete within typical network latency (≤1s).
- The server should limit payload size by excluding unnecessary data.

## Security
- Sanitize all fields before injecting into the DOM to prevent XSS.

## Filtering and Pagination
- `GET /projects/:prid/struct` accepts optional query parameters:
  - `depth` to limit recursion of nested entities.
  - `status` to filter requirements by state (e.g., active, archived).
  - `page` and `page_size` to paginate large requirement lists.

## Real-time Updates
- Provide `GET /projects/:prid/struct/subscribe` using WebSockets or Server-Sent Events.
- Clients update the DOM when notified of changes to projects, requirements, or attachments.

## Access Control
- All project-structure endpoints require authentication.
- Roles `viewer`, `editor`, and `admin` restrict read/write/delete abilities.

## Attachment Management
- Users can add and remove requirement attachments.
  - `POST /requirements/:rid/attachments` uploads a file.
  - `DELETE /requirements/:rid/attachments/:aid` removes an attachment.
- Uploaded attachments are included in project-structure responses.

## Design Views
- Expose `GET /projects/:prid/design` so clients can fetch project design documentation.
- Returned artifacts (e.g., markdown or diagrams) are rendered in the DOM.

## Export Formats
- Support additional exports:
  - `GET /projects/:prid/export/csv`
  - `GET /projects/:prid/export/pdf`
- These endpoints return the same nested structure in CSV or PDF formats.

## Reference
See [architecture.md](architecture.md) for the overall system diagram and endpoint mappings.
