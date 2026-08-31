FROM nginx:1.27-alpine

# Template processado com envsubst na inicialização do container.
COPY nginx/default.conf.template /etc/nginx/templates/default.conf.template

COPY index.html /usr/share/nginx/html/index.html
COPY css/       /usr/share/nginx/html/css/
COPY js/        /usr/share/nginx/html/js/

ENV BACKEND_URL=http://backend:8080

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
    CMD wget -qO- http://127.0.0.1/healthz || exit 1
