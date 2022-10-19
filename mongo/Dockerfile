FROM mongo:4.2
WORKDIR /mongo
COPY . .
RUN chmod 400 key.cert
RUN chown 999:999 key.cert
RUN chmod +x rs_initiate.sh health_check.sh

EXPOSE 27017

HEALTHCHECK --interval=5s --timeout=5s --retries=5 --start-period=5s CMD ./health_check.sh

CMD [ "mongod",  \
  "--replSet=rs0", \
  "--bind_ip_all", \
  "--keyFile=/mongo/key.cert", \
  "--dbpath=/data/db" \
  ]
