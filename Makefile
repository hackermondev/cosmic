run: cosmic
	./builds/cosmic

cosmic: src/*.go
	cd src && go build -o ../builds