all:
	go build -o ./bin/world ./cmd/world/ | go build -o ./bin/sim ./cmd/sim/ | go build -o ./bin/gen ./cmd/gen/
world:
	go build -o ./bin/world ./cmd/world/
sim:
	go build -o ./bin/sim ./cmd/sim/
gen:
	go build -o ./bin/gen ./cmd/gen/
winworld:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/world.exe ./cmd/world
winsim:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/sim.exe ./cmd/sim
wingen:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/gen.exe ./cmd/gen
winall:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/world.exe ./cmd/world/ | go build -o ./bin/sim.exe ./cmd/sim/ | go build -o ./bin/gen.exe ./cmd/gen/
