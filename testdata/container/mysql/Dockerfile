FROM mysql:5

# Create app directory
RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

# Bundle app source
COPY . /usr/src/app
RUN chmod +x /usr/src/app/*.sh

COPY entry-mysql.sh /docker-entrypoint-initdb.d/
