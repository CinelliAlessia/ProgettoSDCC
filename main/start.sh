#!/bin/bash

# Imposta la variabile di ambiente

# 1 locale, 2 docker
export CONFIG=2
# 1=true 0=false
export DEBUG=0


if [ "$CONFIG" -eq 1 ]; then

pkill gnome-terminal
  # Percorsi dei file Go
  server_file="server/"
  client_file="client/"

  # Numero di repliche
  Replicas=3

  # Esecuzione del server in nuovi terminali gnome-terminal
  for ((i=0; i<Replicas; i++)); do
    gnome-terminal --geometry=100x24+0+0 -- bash -c "cd $server_file && go run . $i; exec bash" &
  done

  # Esecuzione del client in un nuovo terminale gnome-terminal
  gnome-terminal --geometry=100x24+0+0 -- bash -c "cd $client_file && go run .; exec bash" &

elif [ "$CONFIG" -eq 2 ]; then
  # Esecuzione con Docker Compose
  #docker-compose up --build
  docker-compose -f compose.yml up --build -d
else
  echo "Configurazione non supportata, inserire 1, 2 o 3."
fi
