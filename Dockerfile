FROM golang:1.18 AS build

ARG RELEASE_STRING=dev
ENV IMPORT_PATH="github.com/plumber-cd/chcontainer"
WORKDIR /go/delivery
COPY . .
RUN mkdir bin && go build \
    -ldflags "-X ${IMPORT_PATH}.Version=${RELEASE_STRING}" \
    -o ./bin ./...

FROM alpine

ARG user=chcontainer
ARG group=chcontainer
ARG uid=1000
ARG gid=1000
ARG home=/home/${user}

RUN addgroup -g ${gid} ${group}
RUN adduser -h ${home} -u ${uid} -S ${user} -G ${group}

COPY --from=build /go/delivery/bin /usr/bin
# COPY chcontainer /usr/bin/chcontainer

USER ${user}
WORKDIR ${home}
CMD ["cat"]
HEALTHCHECK NONE
