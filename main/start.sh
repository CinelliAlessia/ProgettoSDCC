#!/bin/bash

# Imposta la variabile di ambiente

# 1 locale, 2 docker
export CONFIG=1
# 1=true 0=false
export DEBUG=1

pkill gnome-terminal

if [ "$CONFIG" -eq 1 ]; then

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
  docker-compose up --build
else
  echo "Configurazione non supportata, inserire 1 o 2."
fi
