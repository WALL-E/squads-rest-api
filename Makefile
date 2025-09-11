APP_NAME=squads-rest-api

run:
	go mod tidy
	go run main.go

build:
	go mod tidy
	go build -o $(APP_NAME) main.go

clean:
	rm -f $(APP_NAME) app.db
