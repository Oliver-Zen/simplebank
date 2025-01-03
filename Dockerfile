###########################################
### Build Stage: Compile Go Application ###
###########################################
FROM golang:1.23-alpine AS builder
# Using Go 1.23 on Alpine as the [base image] for building
WORKDIR /app
COPY . .
# Copies all local files into /app inside the container

RUN go build -o main main.go
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xvz
        
# build our app to a single binary executable file

#########################################################
### Run Stage: Minimal Alpine Image with Compiled App ###
#########################################################
FROM alpine
# A lightweight image to keep the final container small
WORKDIR /app
COPY --from=builder /app/main .
# Copies only the compiled binary from the builder stage
COPY --from=builder /app/migrate ./migrate
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY db/migration ./migration

# best practice: use EXPOSE to inform Docker that ¬
# the container listens on the specified network port at runtime
EXPOSE 8080
# EXPOSE does not actually publish the port
# only serves as a documentation btw image builder & image runner

CMD [ "/app/main" ]
# The default command that starts the Go application
ENTRYPOINT [ "/app/start.sh" ]
# when CMD instruction is used together with ENTRYPOINT
# CMD acts as just the additional params
# i.e., equivilant to running ENTRYPOINT [ "/app/start.sh", "/app/main" ]