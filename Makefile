all:
	go build -o ./bin/world ./cmd/world/ | go build -o ./bin/sim ./cmd/sim/ | go build -o ./bin/gen ./cmd/gen/
world:
	go build -o ./bin/world ./cmd/world/
phys:
	go build -o ./bin/sim ./cmd/sim/
gen:
	go build -o ./bin/gen ./cmd/gen/
