# go-space-serv
A pair of servers built on [https://github.com/panjf2000/gnet](https://github.com/panjf2000/gnet).
|Program|Description|Location
|--|--|--|
|WORLD|uses TCP to authenticate and communicate data to clients.|`build/unix/world`, `build/win/world.exe`|
|SIM|uses UDP to propagate input and maintain physics authority.|`build/unix/sim`, `build/win/sim.exe`|
|GEN|creates map data.|`build/unix/gen`, `build/win/gen.exe`|

TODO: Architecture diagrams.
## Usage
### Step One -- Generate Map Data

osx: `./build/unix/gen assets/localMap`

windows: `build\win\gen.exe assets/localMap`


```
osx
===
./build/unix/gen --cpf=512 --csize=128 --size=256 --seed=209323094 --threshold=0.36 assets/localMap

windows
=======
build\win\gen.exe --cpf=512 --csize=128 --size=256 --seed=209323094 --threshold=0.36 assets/localMap
```
WORLD and SIM do not have their own flags yet, so they are hardcoded to look for the map in `assets/localMap`.
### Step Two -- Start WORLD
osx: `./build/unix/world`

windows: `build\win\world.exe`

WORLD will load `assets/localMap/meta.chunks` and print it out.

It will then wait for SIM to connect.

### Step Three -- Start SIM
osx: `./build/unix/sim`

windows: `build\win\sim.exe`

SIM will connect via tcp to WORLD.

WORLD will now begin accepting client connections.

Both programs look for map data in `assets/localMap`

When a client connects to WORLD, WORLD passes this information to SIM.

SIM will now accept UDP packets from this IP.

Clients will then perform a virtual connection sequence with SIM.

Once SIM is satisfied it will synchronize this client with the simulation and allow input.

The reliable ordered UDP protocol follows principles from these articles: [https://www.gafferongames.com/](https://www.gafferongames.com/)


### Step Four -- Shutdown
Shut down both servers by terminating the WORLD process.


## Command Syntax
all commands use the same syntax:
`cmd --flag=value argument`

where cmd is the name of the program

one or more flags as specified in the tables below

and an argument

## Building
Building is done using `make`
The makefile is in the root of the project.
```
unix targets
============
make gen
make world
make sim

windows targets
===============
make wingen
make winworld
make winsim
```
## How It Works
### GEN

1) `GEN` generates a simplex noise profile.
2) Each coordinate on the map is tested against this profile and a threshold value to determine
if it is solid or empty.
3) This information is then zipped and saved to a numbered file. Example: `assets/localMap/000.chunks`.
4) information based on flag inputs is saved to `assets/localMap/meta.chunks`


|Flag|Default|Description|
|--|--|--|
|cpf|512|How many chunks to put into one file.|
|csize|128|How many blocks are in a chunk.|
|size|256|Width and height of map in chunks.|
|**seed**|209323094|noise seed.|
|**threshold**|0.36|Threshold value for solid/empty.|
|clean|false|Clean without generating the map.|

argument is a relative location to store the files.

### WORLD
1) Loads the map specified by the argument passed to the command.
2) Block and listen on the specified port for SIM.
3) Allow player connections.
4) On player connection send IP to SIM
5) Send information about SIM to player.
6) Receive updates from SIM about player positions.
7) Send map data to players based on their position.

|Flag|Default|Description|
|--|--|--|
|port|9494|Which port to listen on.|

argument is a relative location to where the map files are stored.

### SIM
1) Begins listening on specified port and tells WORLD to accept connections.
2) Engages clients in handshake and adds them to connected players
3) Run all input through local simulation, validate, and propagate.

|Flag|Default|Description|
|--|--|--|
|cpuprofile|false|whether to generate a cpu profile.|
|timestep|33|How many milliseconds per frame.|
|timestepNano|33000000|How many nanoseconds per frame.|
|protocolId|3551548956|Must be the same on client. Is a hash of project name and version.|
|worldRate|12|How many frames between state updates sent to WORLD|
