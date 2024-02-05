FROM golang:latest

WORKDIR /opt

# Set default OS to linux and architecture to amd64
ENV GOOS=linux
ENV GOARCH=amd64

# mount project in workdir
ADD internal          internal
ADD go.mod            go.mod
ADD go.sum            go.sum
ADD OpenConnectSSO.go OpenConnectSSO.go

# create output dir
RUN mkdir out

# build executable
CMD ["go", "build", "-o", "out/openconnect-sso", "."]