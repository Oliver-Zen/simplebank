########################################################
### write Dockerfile to package the app for shipping ###
########################################################

### define the [base image] to build our app ###
FROM golang:1.23-alpine
WORKDIR /app
COPY . .
# second dot is the current working dir inside the image
# where files and folders are being copied to (`WORKDIR`)


### build our app to a single binary executable file ###
RUN go build -o main main.go

# best practice: use EXPOSE to inform Docker that ¬
# the container listens on the specified network port at runtime
EXPOSE 8080
# EXPOSE does not actually publish the port
# it only serves as a documentation btw image builder & image runner


### define the default command to run when the container starts ###
CMD [ "/app/main" ]

##################################################################
### 500MB <- contains golang and all required packages ###########
### so, covert Dockerfile to [multi-stage], to reduce the size ###
### even avoid the original golang code (just binary) ############
##################################################################