FROM sin.ocir.io/ax4p6mavrtb1/dmdc/golang:1.25 as builder

WORKDIR /app

COPY . .
RUN go mod tidy
RUN go build -o http .

# Distribution

FROM sin.ocir.io/ax4p6mavrtb1/dmdc/golang:1.25
ENV TZ=Asia/Jakarta

WORKDIR /app 
EXPOSE 8080

COPY --from=builder /app/http /app
CMD /app/http
