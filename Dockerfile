FROM golang:1.18
ENV GOOS linux
ENV CGO_ENABLED 0
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o ctf
EXPOSE 5000
ENV DBUSER postgres
ENV DBPASSWORD postgres
ENV DBHOST 127.0.0.1
ENV DBPORT 5432
ENV DBNAME postgres
CMD ./ctf