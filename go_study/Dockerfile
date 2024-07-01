FROM golang:1.22.3 AS build

WORKDIR /app

COPY . /app

RUN go build -o app ./cmd/api/*.go

FROM golang:1.22.3 AS final

WORKDIR /app

COPY --from=build /app /app

EXPOSE 8080

CMD [ "./app" ]
