FROM golang:latest 

RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 
RUN go get -u github.com/aws/aws-sdk-go
RUN go get -u github.com/briandowns/spinner
RUN go build -o /app/main 
CMD ["/app/main"]