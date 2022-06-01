FROM scratch
COPY --from=local/certs:1.0.3 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY bin/prom2kafka-prod .
EXPOSE 9201
ENTRYPOINT ["/prom2kafka-prod"]