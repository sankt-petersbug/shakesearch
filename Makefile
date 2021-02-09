.PHONY: test

test:
	go test -race ./...

start: # start server
	@go run main.go

clean: # removes indexes
	rm -rf shakesearch.bleve

deploy:
	git push heroku master
