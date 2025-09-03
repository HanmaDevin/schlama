install:
	@go install .

run: build
	@./bin/schlama

run_win: build_win
	@./bin/schlama.exe

build: 
	@go build -o bin/schlama .

build_win:
	@go build -o bin/schlama.exe .

clean:
	@rm -r ./bin
	
