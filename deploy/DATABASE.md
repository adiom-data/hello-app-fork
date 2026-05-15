# Tenant Database

The app deploy bundle requests a tenant-local PostgreSQL cluster from
CloudNativePG:

```yaml
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: app-db
```

CloudNativePG creates the underlying StatefulSet, PVCs, Services,
certificates, and application credentials.

The API connects to:

```text
host: app-db-rw
port: 5432
database: app
```

CloudNativePG generates the application credentials in this Secret:

```text
app-db-app
```

The `/api/hello` endpoint creates a small `hello_hits` table if needed and
inserts one row for each request. The response includes the stored hit count
and latest database write time.
