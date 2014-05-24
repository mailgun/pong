FROM ubuntu:12.04

# Let's install go just like Docker (from source).
RUN apt-get update -q
RUN DEBIAN_FRONTEND=noninteractive apt-get install -qy build-essential curl git
RUN curl -s https://go.googlecode.com/files/go1.2.src.tar.gz | tar -v -C /usr/local -xz
RUN cd /usr/local/go/src && ./make.bash --no-clean 2>&1
RUN apt-get -y -q install bzr

# Set up environment variables.
ENV PATH /usr/local/go/bin:$PATH
ENV GOROOT /usr/local/go
ENV GOPATH /home/goworld
ENV PONGPATH /home/goworld/src/github.com/mailgun/pong

RUN echo "clear cache 5"
RUN go get -v -u github.com/gorilla/mux
RUN go get -v -u github.com/mailgun/pong
RUN go install github.com/mailgun/pong
RUN mkdir /opt/pong
RUN cp /home/goworld/bin/pong /opt/pong
RUN cp /home/goworld/src/github.com/mailgun/pong/examples/docker.yaml /opt/pong
CMD /opt/pong/pong -c /opt/pong/docker.yaml
