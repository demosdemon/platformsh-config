# platformsh-config

TODO: everything

## Examples

- [Docker Registry](https://github.com/demosdemon/docker-registry/blob/c671f7adf5a595de7649800aa7b88a10831a933f/config.yml)

```yaml
version: 0.1
log:
  accesslog:
    disabled: true
  level: debug
  formatter: text
  fields:
    branch: {{ env "PLATFORM_BRANCH" }}
    environment: {{ env "PLATFORM_ENVIRONMENT" }}
    project: {{ env "PLATFORM_PROJECT" }}
    service: {{ env "PLATFORM_APPLICATION_NAME" }}
    tree_id: {{ env "PLATFORM_TREE_ID" }}
storage:
  filesystem:
    rootdirectory: {{ env "PLATFORM_DIR" }}/var/lib/registry
auth:
  token:
{{- with json (slice (route "$[?(@.upstream == 'auth')]") "$[0].url") }}
    realm: {{ printf "%s://%s%s" .Scheme .Host .Path }}
{{- end }}
    service: Docker Registry
    issuer: Acme Auth Server
    rootcertbundle: {{ env "PLATFORM_DIR" }}/bundle.crt
http:
  addr: localhost:{{ env "PORT" }}
  net: tcp
  prefix: /
{{- with json (slice (route "$[?(@.upstream == 'registry')]") "$[0].url") }}
  host: {{ printf "%s://%s" .Scheme .Host }}
{{- end }}
  secret: {{ env "PLATFORM_PROJECT_ENTROPY" }}
redis:
{{- with json (rel "$.cache[0]") }}
  addr: {{ printf "%v:%v" .host .port }}
  db: 0
{{- with .password }}
  password: {{ . }}
{{- end }}
{{- end }}
```

- [Docker Registry Auth](https://github.com/demosdemon/docker-registry-auth/blob/317141e553bc6adca2517482cc9cef149b904d01/config.yml)

```yaml
server:
  addr: :{{ env "PORT" }}
token:
  issuer: "Acme Auth Server"
  expiration: 900
  certificate: {{ env "PLATFORM_DIR" }}/server.crt
  key: /tmp/server.key
users:
{{- range $name, $password := (json (var "$.users")) }}
  "{{ $name }}":
    password: {{ bcrypt $password }}
{{- end }}
acl:
  - match: {account: "admin"}
    actions: ["*"]
    comment: Admin has full access to everything
  - match: {account: "/.+/", name: "${account}/*"}
    actions: ["*"]
    comment: Logged in users have full access to images that are in their namespace
  - match: {account: "/.+", type: "registry", name: "catalog"}
    actions: ["*"]
    comments: Logged in users can query the catalog.
  - match: {name: "${labels:project}/*"}
    actions: ["push", "pull"]
    comment: Users can push to any project they are assigned to
# access is denied by default
```
