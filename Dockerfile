
FROM  ubuntu:12.04

MAINTAINER  "John Goodall <jgoodall@ornl.gov>"

# make sure the package repository is up to date
RUN  echo "deb http://archive.ubuntu.com/ubuntu precise main universe" > /etc/apt/sources.list
RUN  apt-get update

# Install ubuntu dependencies
RUN  apt-get install -y make python g++ curl

# Install node.js
RUN   curl http://nodejs.org/dist/v0.10.20/node-v0.10.20-linux-x64.tar.gz | tar -C /usr/local/ --strip-components=1 -zxv

# Add application files
ADD  ./doc-service /doc-service

# Setup node modules, if necessary
# RUN  cd /doc-service && npm install -d

# Expose application port
EXPOSE  8000

# Start the application
CMD  ["cd" "/doc-service;", "npm", "start"]