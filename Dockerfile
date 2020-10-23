FROM harbor.baoding.paas.gwm.cn/library/centos:7.7
COPY ./tyk-pump /opt/tyk-pump/
EXPOSE 8000
ENTRYPOINT ["/opt/tyk-pump/tyk-pump"]
