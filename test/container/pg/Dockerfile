FROM postgres:12.0

# Create app directory
RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

# Bundle app source
COPY . /usr/src/app

COPY 0*.sh /docker-entrypoint-initdb.d/

