#!/bin/bash
#
# Run the Foc engine

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PROJECT_ROOT=$SCRIPT_DIR/..

display_help() {
  echo "Usage: foc-engine run [options...]"
  echo
  echo "Options:"
  echo "  --help, -h           Show this help message"
  echo "  --network <network>  Specify the network to run on (devnet, sepolia, mainnet) [default: devnet]"
  echo "  --config <file>      Foc engine configuration file"
  echo "  --preset <preset>    Use a preset configuration for the engine"
  echo "                       Presets: default - All components"
  echo "                                paymaster - AVNU Paymaster only"
  echo
  echo "  --no-clean           Do not clean up volumes before starting"
  echo "  --no-build           Do not build the Docker image before starting"
  echo
  echo "Examples:"
  echo "  foc-engine run"
  echo "  foc-engine run --network sepolia --preset paymaster"
}

# Transform long options to short options
for arg in "$@"; do
  shift
  case "$arg" in
    "--help") set -- "$@" "-h" ;;
    "--network") set -- "$@" "-n" ;;
    "--config") set -- "$@" "-c" ;;
    "--preset") set -- "$@" "-p" ;;
    "--no-clean") set -- "$@" "-C" ;;
    "--no-build") set -- "$@" "-B" ;;
     --*) unrecognized_options+=("$arg") ;;
     *) set -- "$@" "$arg"
  esac
done

# Check if unknown options are passed, if so exit
if [ ! -z "${unrecognized_options[@]}" ]; then
  echo "Error: invalid option(s) passed ${unrecognized_options[*]}"
  exit 1
fi

# Default values for options
NETWORK="devnet"
PRESET="default"
NO_CLEAN=false
NO_BUILD=false

# Parse command line options
while getopts ":hn:c:p:CB" opt; do
  case $opt in
    h)
      display_help
      exit 0
      ;;
    n)
      NETWORK="$OPTARG"
      ;;
    c)
      CONFIG_FILE="$OPTARG"
      ;;
    p)
      PRESET="$OPTARG"
      ;;
    C)
      NO_CLEAN=true
      ;;
    B)
      NO_BUILD=true
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      exit 1
      ;;
    :)
      echo "Option -$OPTARG requires an argument." >&2
      exit 1
      ;;
  esac
done

# Check if the specified network is valid
if [[ ! "$NETWORK" =~ ^(devnet|sepolia|mainnet)$ ]]; then
  echo "Error: Invalid network specified. Valid options are: devnet, sepolia, mainnet."
  exit 1
fi

# Check if the specified preset is valid
if [[ ! "$PRESET" =~ ^(default|paymaster)$ ]]; then
  echo "Error: Invalid preset specified. Valid options are: default, paymaster."
  exit 1
fi

if [ "$PRESET" == "paymaster" ]; then
  if [ "$NETWORK" == "devnet" ]; then
    echo "Warning: Paymaster preset is not available for devnet. Switching to sepolia network."
    NETWORK="sepolia"
  fi
fi

if [ "$NETWORK" == "devnet" ]; then
  DOCKER_COMPOSE_FILE="$PROJECT_ROOT/docker-compose-devnet.yml"
elif [ "$NETWORK" == "sepolia" ]; then
  DOCKER_COMPOSE_FILE="$PROJECT_ROOT/docker-compose-sepolia.yml"
else
  DOCKER_COMPOSE_FILE="$PROJECT_ROOT/docker-compose-mainnet.yml"
fi

if [ "$NO_BUILD" = false ]; then
  echo "Building Docker image(s) for Foc Engine with preset: $PRESET..."
  if [ "$PRESET" == "paymaster" ]; then
    docker build -t paymaster-image -f $PROJECT_ROOT/dockerfiles/Dockerfile.paymaster $PROJECT_ROOT/.
  else
    docker compose -f "$DOCKER_COMPOSE_FILE" build
  fi
fi

if [ "$NO_CLEAN" = false ]; then
  echo "Cleaning up existing volumes and containers..."
  if [ "$PRESET" == "paymaster" ]; then
    docker container rm paymaster-server
  else
    docker compose -f "$DOCKER_COMPOSE_FILE" down --volumes
  fi
fi

echo "Starting Foc Engine with network: $NETWORK and preset: $PRESET"
if [ "$PRESET" == "paymaster" ]; then
  echo "Running Paymaster server..."
  docker run --rm -p 8080:8080 --name paymaster-server paymaster-image
else
  echo "Running Foc Engine with Docker Compose..."
  docker compose -f "$DOCKER_COMPOSE_FILE" up
fi

# TODO: Use config file
# TODO: Use preset configuration
# TODO: Use an image from Docker Hub or a local build
