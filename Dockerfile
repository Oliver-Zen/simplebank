###########################################
### Build Stage: Compile Go Application ###
###########################################
FROM golang:1.23-alpine AS builder
# Using Go 1.23 on Alpine as the [base image] for building
WORKDIR /app
COPY . .
# Copies all local files into /app inside the container

RUN go build -o main main.go
# build our app to a single binary executable file

#########################################################
### Run Stage: Minimal Alpine Image with Compiled App ###
#########################################################
FROM alpine
# A lightweight image to keep the final container small
WORKDIR /app
COPY --from=builder /app/main .
# Copies only the compiled binary from the builder stage
COPY app.env .

# best practice: use EXPOSE to inform Docker that ¬
# the container listens on the specified network port at runtime
EXPOSE 8080
# EXPOSE does not actually publish the port
# only serves as a documentation btw image builder & image runner

CMD [ "/app/main" ]
# The default command that starts the Go application