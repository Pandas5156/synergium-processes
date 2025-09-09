FROM golang:1.24-alpine

WORKDIR /app

COPY backend/go.mod ./  
COPY backend/go.sum ./

RUN go mod download

COPY . .

CMD ["go", "run", "main.go"]