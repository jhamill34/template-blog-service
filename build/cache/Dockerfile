FROM redis:latest
EXPOSE 6379
CMD redis-server --appendonly yes --requirepass $(cat $CACHE_PASSWORD)
HEALTHCHECK --timeout=10s --retries=10 CMD redis-cli --raw incr ping

