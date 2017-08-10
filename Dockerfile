FROM ubuntu
ADD bin/linux/sdv-linux-x64 sdv-env.sh /

# you won't want to change these as this sets up sdv to listen outside of the docker container
ENV sdvListenOn "0.0.0.0"
ENV sdvPort "8080"

# you'll want to override these with your own
ENV sdvDriver "sqlite"
ENV sdvDb "/eg/Chinook_Sqlite_AutoIncrementPKs.sqlite"

# embed a test database so it works out of the box
ADD db/Chinook_Sqlite_AutoIncrementPKs.sqlite /eg/

CMD ["/sdv-env.sh"]

EXPOSE 8080
