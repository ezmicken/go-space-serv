all:
	go build -o ./bin/world ./cmd/world/ | go build -o ./bin/sim ./cmd/sim/
world:
	go build -o ./bin/world ./cmd/world/
phys:
	go build -o ./bin/sim ./cmd/sim/
