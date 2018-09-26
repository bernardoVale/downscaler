FROM scratch

COPY ./downscaler /downscaler

EXPOSE 8080
ENV DEBUG=true

ENTRYPOINT [ "/downscaler" ]