FROM docker.io/library/postgres:15-alpine

COPY --chown=postgres:postgres postgres-backup.sh /usr/local/bin/postgres-backup
RUN chmod 0755 /usr/local/bin/postgres-backup \
    && mkdir -p /backups \
    && chown postgres:postgres /backups

USER postgres
ENTRYPOINT ["/usr/local/bin/postgres-backup"]
