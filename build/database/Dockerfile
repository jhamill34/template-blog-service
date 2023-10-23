FROM mysql:latest
EXPOSE 3306
COPY ./build/database/db /docker-entrypoint-initdb.d/
HEALTHCHECK --timeout=10s --retries=10 CMD mysqladmin ping -h localhost

