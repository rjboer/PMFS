# Web Interface User Story

## View Project Structure

As a project manager, I want to retrieve the structure of a project via the web interface so that I can examine the project's requirements. The endpoint `GET /projects/:prid/struct` supports the following optional query parameters to refine the result:

- `depth` – limit the maximum depth of nested requirements.
- `status` – filter requirements by their status (e.g. active, deleted).
- `page` – select a page of paginated results.

When no query parameters are provided, the entire project structure is returned.
