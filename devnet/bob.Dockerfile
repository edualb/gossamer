# Copyright 2021 ChainSafe Systems (ON)
# SPDX-License-Identifier: LGPL-3.0-only

FROM golang:1.18beta1

ARG DD_API_KEY=somekey
ENV DD_API_KEY=${DD_API_KEY}
RUN DD_AGENT_MAJOR_VERSION=7 DD_INSTALL_ONLY=true DD_SITE="datadoghq.com" bash -c "$(curl -L https://s3.amazonaws.com/dd-agent/scripts/install_script.sh)"

WORKDIR /gossamer

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install -trimpath github.com/ChainSafe/gossamer/cmd/gossamer

# use modified genesis-spec.json with only 3 authority nodes
RUN cp -f devnet/chain/gssmr/genesis-spec.json chain/gssmr/genesis-spec.json

ARG key
RUN test -n "$key"
ENV key=${key}

RUN gossamer --key=${key} init

ARG METRICS_NAMESPACE=gossamer.local.devnet

WORKDIR /gossamer/devnet

RUN go run cmd/update-dd-agent-confd/main.go -n=${METRICS_NAMESPACE} -t=key:${key} > /etc/datadog-agent/conf.d/openmetrics.d/conf.yaml

ENTRYPOINT service datadog-agent start && gossamer --key=${key} --bootnodes=/dns/alice/tcp/7001/p2p/12D3KooWMER5iow67nScpWeVqEiRRx59PJ3xMMAYPTACYPRQbbWU --publish-metrics --rpc --pubdns=${key}

EXPOSE 7001/tcp 8545/tcp 8546/tcp 8540/tcp 9876/tcp
