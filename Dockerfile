FROM golang as builder

COPY / /anno
RUN cd /anno && go build \
    -a \
    -installsuffix nocgo \
    -mod=readonly \
    -o /bin/anno \
    /anno/main.go

FROM ubuntu

COPY --from=builder /bin/anno /bin/anno
RUN groupadd --system anno --gid 999 && \
    useradd --no-log-init --system --gid anno anno --uid 999

USER 999:999

CMD [ "/bin/anno" ]
