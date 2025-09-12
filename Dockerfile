FROM golang:1.23.4 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o forum .

FROM debian:bookworm-slim
LABEL project="forum" \
      authors="Ali Isa, Hussain Ali, Hussain Alnasser, Ebrahim Alnasser"
WORKDIR /app
COPY --from=builder /app/forum .
COPY --from=builder /app/Frontend ./Frontend
COPY --from=builder /app/static ./static
COPY --from=builder /app/Backend ./Backend
COPY --from=builder /app/backgrounds ./backgrounds
COPY --from=builder /app/forum.db ./forum.db
EXPOSE 8888
CMD ["./forum"]
