FROM ubuntu
# MAINTAINER timabell
# ADD https://www.dropbox.com/s/5u191ybq3tzfygt/sdv-linux-x64?dl=1 /sdv
ADD bin/linux/sdv-linux-x64 sdv-env.sh /

# you won't want to change these as this sets up sdv to listen outside of the docker container
ENV sdvListenOn "0.0.0.0"
ENV sdvPort "8080"

# you'll want to override these with your own
ENV sdvDriver "mssql"
ENV sdvDb "server=sdv-adventureworks.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=AdventureWorksLT"

CMD ["/sdv-env.sh"]

EXPOSE 8080
