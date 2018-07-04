FROM golang




COPY . /lain/app/ 


WORKDIR /lain/app/


RUN ( apt-get -y update ) && ( apt-get -y install unzip ) && ( curl -sL https://deb.nodesource.com/setup_9.x | su --preserve-environment - ) && ( apt-get install -y nodejs ) && ( go get github.com/tools/godep ) && ( apt-get -y install libldap-dev ) && ( mkdir -p /go/src/github.com/laincloud/ ) && ( ln -sf /lain/app /go/src/github.com/laincloud/sso-ldap ) && ( go get github.com/mijia/gobuildweb ) && ( cd $GOPATH/src/github.com/mijia/gobuildweb && sed -i '/deps = append(deps, "browserify", "coffeeify", "envify", "uglifyify", "babelify", "babel-preset-es2015", "babel-preset-react", "nib", "stylus")/d' cmds.go  && go install ) && ( cd /go/src/github.com/laincloud/sso-ldap && make go-build ) && ( ls -1 | grep -v '\bnode_modules\b' | xargs rm -rf )

