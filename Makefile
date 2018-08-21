build: go-build

go-build:
	gobuildweb dist

go-dep-save:
	dep ensure

test:
	TEST_MYSQL_DSN="test:test@(x.x.x.x:3306)/sso_test" go test -p 1 ./...

.PHONY: build go-build go-dep-save test
