build:
	docker build -t park .

run: build
	docker run --rm -p 5000:5000 park

