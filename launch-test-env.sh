#!/usr/bin/env bash

set -eu

COMPOSE_FLAGS=(-f ci-compose.yaml)
ASH_HOST="http://127.0.0.1:7377"

function print_usage() {
  echo "Usage: ./launch-test-env.sh [OPTION]"
  echo "OPTION:"
  echo "   -c          clean the environment (stop processes / kill docker) "
  echo "   -W          complete wipeout of the environment"
  echo "   -t          run tests..."
  echo "   -b          skip rebuild"
}


function cleanup() {
  echo "========= CLEANING UP COMPOSE ==========="
  docker compose "${COMPOSE_FLAGS[@]}" down
}

function wipeout() {
  echo "========= COMPLETE WIPEOUT OF ENV ==========="
  docker compose "${COMPOSE_FLAGS[@]}" down --rmi 'all'
  sudo rm -rf cdn
  rm -rf test*
}

trap 'cleanup' SIGINT SIGTERM

t_flag=''
b_flag=''
while getopts 'htcWb' flag; do
  case "${flag}" in
    b) b_flag='true';;
    t) t_flag='true';;
    c) cleanup
       exit 0;;
    W) wipeout
       exit 0;;
    h) print_usage
       exit 0 ;;
    *) print_usage
       exit 1 ;;
  esac
done

echo "========= BUILDING PROJECT ==========="
if ! [[ "${b_flag}" == "true" ]]; then
  docker compose "${COMPOSE_FLAGS[@]}" build --no-cache
fi
docker compose "${COMPOSE_FLAGS[@]}" up -d

if [[ "$t_flag" == "true" ]]; then
  echo "========= RUNNING TESTS ==========="
  for file in ./shadows/* ; do
    printf " TESTING: %s" "${file}..."
    cdn_file=$(curl -ksi -X POST -H "Content-Type: multipart/form-data" -F "data=@${file}" "${ASH_HOST}"/ash/upload | tail -n1)
    curl -ks "${ASH_HOST}"/"${cdn_file}" -o test
    diff test "${file}"
    echo " PASSED"    
  done

  cleanup
fi

