FROM mysql:8.0

# # Create app directory
# RUN mkdir -p /usr/src/app
# WORKDIR /usr/src/app

# # Bundle app source
# COPY . /usr/src/app

COPY *.sql /docker-entrypoint-initdb.d/
