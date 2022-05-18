FROM python:3.10.4-slim-buster

ARG TERRAFORM_VERSION=1.0.7
ARG KUBE_VERSION=1.21.1
ARG HELM_VERSION=3.6.0
ARG STEP_VERSION=0.17.2
ARG VELERO_VERSION=1.7.0
ARG ARGOCD_VERSION=2.3.1
ARG YQ_VERSION=4.23.1
ARG MINIO_CLI_VERSION=RELEASE.2022-03-31T04-55-30Z
ARG GITHUB_CLI_VERSION=2.8.0
ARG SVU_VERSION=1.9.0
ARG GORELEASER_VERSION=1.8.3

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

RUN apt-get update &&\
  apt-get -y --no-install-recommends install \
  ca-certificates=20200601~deb10u2 \
  curl=7.64.0-4+deb10u2 \
  git=1:2.20.1-2+deb10u3 \
  jq=1.5+dfsg-2+b1 \
  gettext=0.19.8.1-9 \
  python3-pip=18.1-5 \
  libssl-dev \
  libffi-dev=3.2.1-9 \
  python-dev=2.7.16-1 \
  rsync=3.1.3-6 \
  openssh-client=1:7.9p1-10+deb10u2 \
  sshpass=1.06-1 \
  unzip=6.0-23+deb10u2 \
  wget=1.20.1-1.1 \
  less=487-0.1+b1 \
  mtr=0.92-2 \
  build-essential=12.6 && \
  openssh-client && \
  rm -rf /var/lib/apt/lists/* &&\
  apt-get clean

WORKDIR /tmp

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN wget -q -O "/usr/local/bin/kubectl" "https://storage.googleapis.com/kubernetes-release/release/v${KUBE_VERSION}/bin/$TARGETOS/$TARGETARCH/kubectl" && \
  chmod +x /usr/local/bin/kubectl && \
  wget -q "https://get.helm.sh/helm-v${HELM_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz" -O - | tar -xzO $TARGETOS-$TARGETARCH/helm > /usr/local/bin/helm && \
  chmod +x /usr/local/bin/helm && \
  helm plugin install https://github.com/databus23/helm-diff  && \
  wget -q  "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_${TARGETOS}_${TARGETARCH}.zip" && \
  unzip terraform_${TERRAFORM_VERSION}_${TARGETOS}_${TARGETARCH}.zip -d /usr/bin && \
  wget -q  "https://github.com/smallstep/cli/releases/download/v${STEP_VERSION}/step_${TARGETOS}_${STEP_VERSION}_${TARGETARCH}.tar.gz" && \
  tar xvzf step_${TARGETOS}_${STEP_VERSION}_${TARGETARCH}.tar.gz && \
  mv step_${STEP_VERSION}/bin/step /usr/local/bin/step && \
  chmod +x /usr/local/bin/step && \
  wget -q  "https://github.com/vmware-tanzu/velero/releases/download/v${VELERO_VERSION}/velero-v${VELERO_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz" && \
  tar xvzf velero-v${VELERO_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz && \
  mv velero-v${VELERO_VERSION}-${TARGETOS}-${TARGETARCH}/velero /usr/local/bin/velero && \
  chmod +x /usr/local/bin/velero && \
  wget -q  "https://github.com/caarlos0/svu/releases/download/v${SVU_VERSION}/svu_${SVU_VERSION}_${TARGETOS}_${TARGETARCH}.tar.gz" && \
  tar xvzf svu_${SVU_VERSION}_${TARGETOS}_${TARGETARCH}.tar.gz && \
  mv svu /usr/local/bin/svu && \
  chmod +x /usr/local/bin/svu && \
  wget -q  "https://github.com/argoproj/argo-cd/releases/download/v${ARGOCD_VERSION}/argocd-${TARGETOS}-${TARGETARCH}" -O /usr/local/bin/argocd && \
  chmod +x /usr/local/bin/argocd && \
  wget -q "https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/yq_${TARGETOS}_${TARGETARCH}" -O /usr/local/bin/yq && \
  chmod +x /usr/local/bin/yq && \
  wget -q "https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VERSION}/goreleaser_${GORELEASER_VERSION}_${TARGETARCH}.deb" && \
  dpkg -i goreleaser_${GORELEASER_VERSION}_${TARGETARCH}.deb && \
  wget -q "https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/yq_${TARGETOS}_${TARGETARCH}" -O /usr/local/bin/yq && \
  chmod +x /usr/local/bin/yq && \
  wget  -q "https://dl.min.io/client/mc/release/${TARGETOS}-${TARGETARCH}/archive/mc.${MINIO_CLI_VERSION}" -O /usr/local/bin/mc && \
  chmod +x /usr/local/bin/mc  && \
  wget -q "https://github.com/cli/cli/releases/download/v${GITHUB_CLI_VERSION}/gh_${GITHUB_CLI_VERSION}_${TARGETOS}_${TARGETARCH}.deb" && \
  dpkg -i gh_${GITHUB_CLI_VERSION}_${TARGETOS}_${TARGETARCH}.deb && \
  rm -rf /tmp/*

COPY python-requirements.txt python-requirements.txt

RUN pip install --no-cache-dir pip==21.2.4 cffi==1.14.6 && \
  pip install --no-cache-dir -r python-requirements.txt && \
  mkdir -p /ansible /etc/ansible /polycrate /workspace && \
  echo 'localhost' > /etc/ansible/hosts

COPY ansible.cfg /etc/ansible/ansible.cfg

COPY ansible-requirements.yml ansible-requirements.yml

RUN mkdir -p /etc/ansible/roles /etc/ansible/collections && \
  ansible-galaxy collection install --collections-path /etc/ansible/collections community.general && \
  ansible-galaxy install --roles-path /etc/ansible/roles -r ansible-requirements.yml

WORKDIR /polycrate

# COPY . /polycrate

RUN mv cli/${GOOS}-${GOARCH}/polycrate-${GOOS}-${GOARCH} /usr/local/bin/polycrate && \
  chmod +x /usr/local/bin/polycrate && \
  rm -rf cli

ARG APP_BUILD_DATE
ARG APP_BUILD_VERSION
ENV APP_BUILD_DATE=$APP_BUILD_DATE
ENV APP_BUILD_VERSION=$APP_BUILD_VERSION

CMD [ "echo", "Hello. This is Polycrate version ${APP_BUILD_VERSION}, built on ${APP_BUILD_DATE}. This image is best used with the Polycrate CLI: https://accelerator.ayedo.de/polycrate" ]