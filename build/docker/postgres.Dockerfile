FROM postgres

COPY infra/db/init/ /docker-entrypoint-initdb.d/

EXPOSE 5432