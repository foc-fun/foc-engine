<p align="center" style="margin-bottom: 0px !important">
  <img height="250" src="resources/logo.png" alt="foc_engine_logo" align="center">
</p>

<h1 align="center" style="margin-top: 12px !important">FOC Engine</h1>
<h3 align="center" style="margin-top: 4px !important">Let's make Starknet magic!</h3>

## Dependencies

The following dependencies must be installed to run the foc-engine:
- docker
- docker compose
- cmdline tools: `jq`, `yq`

## Install

Install using [asdf](https://asdf-vm.com/):
```
asdf plugin add foc-engine https://github.com/b-j-roberts/asdf-foc-engine.git
asdf install foc-engine latest
asdf global foc-engine latest
```

or clone the repo and build docker compose images:
```
git clone git@github.com:b-j-roberts/foc-engine.git
cd foc-engine
docker compose -f docker-compose-devnet.yml build
```

## Running

If you installed the asdf plugin:
```
foc-engine run
```

### Run from local:
if you cloned the repo:
```
docker compose -f docker-compose-devnet.yml up
```

To re-run with a fresh state/build do the following:
```
docker compose -f docker-compose-devnet.yml down --volumes
docker compose -f docker-compose-devnet.yml build
docker compose -f docker-compose-devnet.yml up
```
