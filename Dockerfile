FROM golang:1.18 as build

WORKDIR /go/src
RUN git clone -b pluggable-components-v2 https://github.com/johnewart/dapr.git 
WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/app

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian11
COPY --from=build /go/bin/app /
CMD ["/app"]

