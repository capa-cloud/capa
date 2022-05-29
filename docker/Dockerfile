# A simplified docker file for building capa base image

FROM centos:centos7

COPY ./capa /runtime/

WORKDIR /runtime

RUN chmod +x  /runtime/capa

ENTRYPOINT ["/runtime/capa"]
