# PMFS Web Interface

This example has been extended into a small web tool for browsing products,
projects and requirements.

```bash
go run examples/webinterface/main.go -dir <database> -addr :8080
```

Open <http://localhost:8080> in a browser. The page uses the REST endpoints
exposed by the server and requires the `X-Role` header. The demo interface uses
role `viewer` for all requests.
