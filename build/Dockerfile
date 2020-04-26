FROM registry.access.redhat.com/ubi8/ubi-minimal

RUN microdnf update -y && \
    microdnf install -y shadow-utils && \
    rm -rf /var/cache/yum && \
    microdnf clean all

RUN groupadd --gid 1000 sops-operator && \
    useradd --uid 1000 --gid 1000 sops-operator

USER sops-operator
WORKDIR /home/sops-operator

COPY sops-operator /usr/local/bin/

ENTRYPOINT ["sops-operator"]
