FROM iron/base

COPY salesforceclient /opt/service/

WORKDIR /opt/service

RUN chmod +x /opt/service/salesforceclient

EXPOSE 8080:8080

CMD /opt/service/salesforceclient