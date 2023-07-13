FROM gcr.io/distroless/static
COPY pop /usr/local/bin/pop
ENTRYPOINT [ "/usr/local/bin/pop" ]
