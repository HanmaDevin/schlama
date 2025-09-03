install:
	@go install .

run: build
	@go run .

build: 
	@go build -o bin/ .

clean:
	@rm -r ./bin
	
