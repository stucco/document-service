# Uses nodejs image http://dockerfile.github.io/#/nodejs

FROM dockerfile/nodejs
MAINTAINER John Goodall <jgoodall@ornl.gov>

RUN apt-get install -y python make g++
RUN npm install -g nodemon

ADD . /doc-service

RUN cd /doc-service && npm install -d

CMD ["nodemon" "/doc-service/server.js"]
