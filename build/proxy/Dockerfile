FROM nginx:latest

COPY ./build/proxy/nginx.conf /etc/nginx/templates/nginx.conf.template
COPY ./keys/_wildcard.joshrasmussen.dev-key.pem /certs/private.pem
COPY ./keys/_wildcard.joshrasmussen.dev.pem /certs/public.pem

EXPOSE 443

