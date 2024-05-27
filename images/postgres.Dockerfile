FROM postgres:alpine

ENV POSTGRES_PASSWORD mysecretpassword
ADD images/provisioning_data.sh /docker-entrypoint-initdb.d/init-db.sh
