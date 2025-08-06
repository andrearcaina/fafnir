FROM postgres

COPY infra/postgres/init/ /docker-entrypoint-initdb.d/

EXPOSE 5432