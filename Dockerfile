FROM golang:alpine
COPY . .
RUN go build -o main main.go
RUN ./main

EXPOSE 80
EXPOSE 443