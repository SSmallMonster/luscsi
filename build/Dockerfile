FROM rockylinux:8

ARG TARGETARCH

COPY ./_output/luscsi.${TARGETARCH} /luscsi

ENTRYPOINT [ "/luscsi" ]