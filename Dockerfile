FROM golang:latest

WORKDIR /opt

# mount project in workdir
ADD ./ ./

# create output dir
RUN mkdir out

# build executable
CMD ["go", "build", "-o", "out/go-openconnect-sso", "."]