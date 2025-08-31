APP_NAME=squads-rest-api

run:
	go run main.go

build:
	go build -o $(APP_NAME) main.go

clean:
	rm -f $(APP_NAME) app.db
