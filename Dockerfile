FROM tdlib:latest

ENV GO_VERSION=1.24.2

RUN apt-get update && apt-get install -y wget zsh docker
RUN sh -c "$(wget https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh -O -)" && \
    chsh -s /bin/zsh

RUN wget -P /tmp "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz" && \
    tar -C /usr/local -xzf "/tmp/go${GO_VERSION}.linux-amd64.tar.gz" && \
    rm "/tmp/go${GO_VERSION}.linux-amd64.tar.gz"

ENV GOPATH=/root/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

RUN go install github.com/vektra/mockery/v2@v2.53.3
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.2
RUN go install github.com/mailru/easyjson/...@latest
# RUN go install github.com/goreleaser/goreleaser/v2@v2.8.2

