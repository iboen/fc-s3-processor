FROM golang:1.20 as build
WORKDIR /app
# Copy dependencies list
COPY go.mod go.sum ./
# Build with optional lambda.norpc tag
COPY generate-content/main.go .
RUN CGO_ENABLED=0 go build -tags lambda.norpc -o main main.go
# Copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2
COPY --from=build /app/main ./main
ENTRYPOINT [ "./main" ]