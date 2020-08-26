FROM golang as builder

COPY / /anno
RUN cd /anno && make

FROM ubuntu

COPY --from=builder /anno/bin/anno-linux /bin/anno
RUN groupadd --system anno --gid 999 && \
    useradd --no-log-init --system --gid anno anno --uid 999

USER 999:999

CMD [ "/bin/anno" ]
