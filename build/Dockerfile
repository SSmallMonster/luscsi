FROM rockylinux:8

RUN yum install nss -y

RUN if [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
    COPY ./_output/luscsi.arm64 /luscsi; \
elif [ "$TARGETPLATFORM" = "linux/amd64" ]; then \
    COPY ./_output/luscsi.amd64 /luscsi; \
fi

ENTRYPOINT [ "/luscsi" ]