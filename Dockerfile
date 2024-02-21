#latest
FROM golang

WORKDIR /app

COPY . .

#Build
WORKDIR /app/bin
RUN go mod download
RUN GOOS=linux go build -o gannett-grocery

EXPOSE 8081

#Run
 ENTRYPOINT ["./gannett-grocery"]