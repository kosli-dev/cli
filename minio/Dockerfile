FROM cgr.dev/chainguard/minio-client:latest AS mc

FROM cgr.dev/chainguard/minio:latest-dev
COPY --from=mc /usr/bin/mc /usr/bin/mc
