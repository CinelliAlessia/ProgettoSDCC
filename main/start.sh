#!/bin/bash

# Imposta la variabile di ambiente
export CONFIG=1 # 1 locale, 2 docker
export DEBUG=1 # 1=true 0=false

pkill gnome-terminal

if [ "$CONFIG" -eq 1 ]; then

  # Percorsi dei file Go
  server_file="server/"
  client_file="client/"

  # Esecuzione del server in un nuovo terminale gnome-terminal
  gnome-terminal --geometry=100x24+0+0 -- bash -c "cd server/ && go run . 0; exec bash" &
  gnome-terminal --geometry=100x24+0+0 -- bash -c "cd $server_file && go run . 1; exec bash" &
  gnome-terminal --geometry=100x24+0+0 -- bash -c "cd $server_file && go run . 2; exec bash" &

  # Esecuzione del client in un nuovo terminale gnome-terminal
  gnome-terminal --geometry=100x24+0+0 -- bash -c "cd $client_file && go run .; exec bash" &

elif [ "$CONFIG" -eq 2 ]; then
  # Esecuzione con Docker Compose
  docker-compose up --build
else
  echo "Configurazione non supportata, inserire 1 o 2."
fi
