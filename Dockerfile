FROM ubuntu:noble 
# 24.04 LTS

# ARG GO_VERSION
# ENV GO_VERSION=${GO_VERSION}

ENV GO_VERSION=1.24.2

RUN apt-get update

RUN apt-get install -y make git zlib1g-dev libssl-dev gperf php-cli cmake clang-18 libc++-18-dev libc++abi-18-dev
RUN git clone https://github.com/tdlib/td.git && \
  cd td && \
  rm -rf build && \
  mkdir build && \
  cd build && \
  CXXFLAGS="-stdlib=libc++" CC=/usr/bin/clang-18 CXX=/usr/bin/clang++-18 cmake -DCMAKE_BUILD_TYPE=Debug -DCMAKE_INSTALL_PREFIX:PATH=../tdlib .. && \
  cmake --build . --target install

RUN apt-get install -y wget git gcc zsh

RUN sh -c "$(wget https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh -O -)"
RUN chsh -s /bin/zsh

RUN wget -P /tmp "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz"
RUN tar -C /usr/local -xzf "/tmp/go${GO_VERSION}.linux-amd64.tar.gz"
RUN rm "/tmp/go${GO_VERSION}.linux-amd64.tar.gz"

ENV GOPATH=/root/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

RUN go install github.com/vektra/mockery/v2@v2.53.3
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.2
RUN go install github.com/goreleaser/goreleaser/v2@v2.8.2


