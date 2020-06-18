FROM kong/go-plugin-tool:2.0.4-alpine-1 as builder

WORKDIR /tmp

# compile go-pluginserver here to make sure package versions match compiled plugins below: https://github.com/Kong/docker-kong/issues/326
RUN git clone -b v0.3.3  --depth 1 https://github.com/Kong/go-pluginserver.git && \
    cd go-pluginserver && \
    make

COPY plugins/ /tmp/kong-plugins/
RUN cd /tmp/kong-plugins/ && go build -buildmode=plugin token-introspection.go

FROM kong:2.0.4-alpine
COPY --from=builder /tmp/go-pluginserver/go-pluginserver /usr/local/bin/go-pluginserver
COPY --from=builder /tmp/kong-plugins/*.so /usr/local/go-plugins/
