all:
	go build -o ./bin/world ./cmd/world/ | go build -o ./bin/phys ./cmd/phys/
world:
	go build -o ./bin/world ./cmd/world/
phys:
	go build -o ./bin/phys ./cmd/phys/
