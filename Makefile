world:
	env GOOS=linux GOARCH=amd64 go build -o ./bin/world ./cmd/world/
sim:
	env GOOS=linux GOARCH=amd64 go build -o ./bin/sim ./cmd/sim/
gen:
	env GOOS=linux GOARCH=amd64 go build -o ./bin/gen ./cmd/gen/
spacesim-osx:
	env GOOS=darwin go build -x -v -o ./bin/spacesim.dylib -buildmode=c-shared ./cmd/spacesim
winworld:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/world.exe ./cmd/world
winsim:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/sim.exe ./cmd/sim
wingen:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/gen.exe ./cmd/gen
winall:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/world.exe ./cmd/world/ | go build -o ./bin/sim.exe ./cmd/sim/ | go build -o ./bin/gen.exe ./cmd/gen/
