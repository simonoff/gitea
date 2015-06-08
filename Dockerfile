FROM google/golang:latest

ENV TAGS="sqlite redis memcache cert" USER="git" HOME="/home/git"

COPY  . /gopath/src/github.com/go-gitea/gitea/
WORKDIR /gopath/src/github.com/go-gitea/gitea/

RUN  go get -v -tags="$TAGS" github.com/go-gitea/gitea \
  && go build -tags="$TAGS" \
  && useradd -d $HOME -m $USER \
  && chown -R $USER .

USER $USER

ENTRYPOINT [ "./gogs" ]

CMD [ "web" ]
