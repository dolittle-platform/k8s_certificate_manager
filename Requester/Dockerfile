### Build image ###
FROM golang:1.11 as build

WORKDIR /go/src/github.com/dolittle-platform/k8s_certificate_manager/Requester/Source

# Install glide
RUN curl https://glide.sh/get | sh
# Install dependencies using glide
COPY Source/glide.* ./
RUN glide install
# Build standalone executable
COPY Source/* ./
RUN CGO_ENABLED=0 go build -a main.go


### Runtime image ###
FROM scratch
COPY --from=build /go/src/github.com/dolittle-platform/k8s_certificate_manager/Requester/Source/main /bin/main
ENTRYPOINT [ "/bin/main" ]